package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"gitlab.bbdev.team/vh/vh-srv-profile/utils"
)

type MessageAndSuccess struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type order struct {
	ID           int       `json:"ID"`
	Status       string    `json:"Status"`
	ProductType  string    `json:"ProductType"`
	PaymentDate  time.Time `json:"PaymentDate"`
	Type         string    `json:"type"`
	Quantity     int       `json:"Quantity"`
	StartingDate time.Time `json:"StartingDate"`
}

type payment struct {
	ID            int       `json:"ID"`
	Amount        int       `json:"Amount"`
	DebitCurrency string    `json:"DebitCurrency"`
	CCNumber      string    `json:"CCNumber"`
	PaymentStatus string    `json:"PaymentStatus"`
	CreatedAt     time.Time `json:"created_at"`
	Status        string    `json:"Status"`
}

type special struct {
	Email       string `json:"email"`
	Category    string `json:"category"`
	SubCategory string `json:"subcategory"`
}

type specialRes struct {
	MessageAndSuccess
	Data special `json:"data"`
}

type paymentRes struct {
	MessageAndSuccess
	Data payment `json:"data"`
}

type multiplePaymentRes struct {
	MessageAndSuccess
	Data []payment `json:"data"`
}

type orderRes struct {
	MessageAndSuccess
	Data []order `json:"data"`
}

type orderDeleteRes struct {
	MessageAndSuccess
	Data int `json:"data"`
}

// TODO: test all the scenarios
// TODO: comment every step of the process
// TODO: check the execution sequence
func (db *pgProfileDB) evaluateMembershipByUserID(ctx context.Context, evalBody emailKeycloakAndUserIDBody, authHeader string) (userMembershipRes, error) {

	var email string
	var userID string
	var userKeycloakID string

	// fetch email, userID and userKeycloakID from the database if not available in the request body
	if evalBody.UserID != nil && *evalBody.UserID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, keycloak_id FROM users WHERE user_id=$1`, *evalBody.UserID).Scan(&email, &userKeycloakID); err != nil {
			return userMembershipRes{}, fmt.Errorf("problem getting email from user_id: %w", err)
		}
		userID = *evalBody.UserID
	} else if evalBody.KeycloakID != nil && *evalBody.KeycloakID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, user_id FROM users WHERE keycloak_id=$1`, *evalBody.KeycloakID).Scan(&email, &userID); err != nil {
			return userMembershipRes{}, fmt.Errorf("problem getting email from keycloak_id: %w", err)
		}
		userKeycloakID = *evalBody.KeycloakID
	} else {
		if err := db.QueryRow(ctx, `SELECT user_id, keycloak_id FROM users WHERE primary_email=$1`, *evalBody.Email).Scan(&userID, &userKeycloakID); err != nil {
			return userMembershipRes{}, fmt.Errorf("problem getting user_id from email: %w", err)
		}
		email = *evalBody.Email
	}

	var membershipInsertData membership
	var currentMembership string
	var latestGrantCreatedAt *time.Time
	var latestOrderPaymentDate *time.Time
	var latestOrderPaymentID *int
	var latestOrderPaymentStatus string
	var grantID *int
	var latestOrder order
	var grantMonthsGranted *int
	var allOrderCancelled bool
	var latestGrantCancelled bool
	var userInSpecialTable bool
	var latestRequest requestResponse

	const LIMIT = 200
	// limit to string

	var orderStartingDate time.Time
	var previousStartingDate time.Time
	var previousOrderQuantity int
	currentMonth := int(time.Now().Month())
	currentYear := time.Now().Year()

	*membershipInsertData.Month = currentMonth
	*membershipInsertData.Year = currentYear

	// if not email exit
	if email == "" {
		return userMembershipRes{}, fmt.Errorf("no email found")
	}

	// check if user has any active request
	userRequests, userRequestsErr := db.getMultipleRequest(ctx, 0, LIMIT, userKeycloakID, "", "", "hhmembership", "desc")

	if userRequestsErr != nil {
		return userMembershipRes{}, fmt.Errorf("problem getting user requests: %w", userRequestsErr)
	}

	if len(userRequests) != 0 {
		// latest request of the user
		latestRequest = userRequests[0]
	}

	// fetch all the order of the user ( limit 200 as of now )
	userOrderDetails := getServerUrl() + "/pay/v2/orders?email=" + email + "&product-type=globalmembership&limit=" + strconv.Itoa(LIMIT) + "&evaluate-membership=true&o-payment-date=desc"

	orderDetails, _ := utils.HTTPCallAndGetBody(userOrderDetails, authHeader, nil, "GET")

	// unmarshal order details
	var orderDetailsRes orderRes
	err := json.Unmarshal(orderDetails, &orderDetailsRes)
	if err != nil {
		fmt.Println("problem unmarshalling order details: %w", err)
	}

	// fetch user grant if any
	userGrant, userGrantErr := db.getMultipleGrant(ctx, 0, LIMIT, nil, userID, "helphaver", "desc")

	if userGrantErr != nil {
		if !errors.Is(userGrantErr, errUserNotFound) {
			fmt.Println("problem getting user grant: %w", userGrantErr)
		}
	}

	if len(userGrant) != 0 {
		// access the latest grant of the user
		latestGrant := userGrant[0]
		grantID = latestGrant.ID

		// check if the latest grant is cancelled to check if the user membership is cancellled
		if latestGrant.CancelledAt != nil {
			latestGrantCancelled = true
		}

		// fetch grant membership based on grant ID
		grantMemb, grantMembErr := db.getGrantMembershipByGrantID(ctx, *grantID)

		if grantMembErr != nil {
			return userMembershipRes{}, fmt.Errorf("problem getting grant membership: %w", grantMembErr)
		}

		// set number of months granted to the user
		grantMonthsGranted = grantMemb.Month

		// created at value of the parent grant table
		latestGrantCreatedAt = latestGrant.CreatedAt
	}

	// Evaluation starts here
	if len(orderDetailsRes.Data) != 0 {

		// count number of cancelled order in orderDetailsRes.Data
		var cancelledOrderCount int
		for _, order := range orderDetailsRes.Data {
			if order.Status == "cancelled" {
				cancelledOrderCount++
			}
		}

		// if all orders are cancelled, then current membership is cancelled
		if cancelledOrderCount != 0 && cancelledOrderCount == len(orderDetailsRes.Data) {
			allOrderCancelled = true
		} else {
			// latest order
			latestOrder = orderDetailsRes.Data[0]
			latestOrderPaymentDate = &latestOrder.PaymentDate

			if latestOrder.Type == "recurring" {
				currentMembership = "automatic"
			} else {
				currentMembership = "manual"
			}

			// fetch latest order payment id
			paymentDetails, _ := utils.HTTPCallAndGetBody(getServerUrl()+"/pay/v2/payments?o-created-at=desc&order-id="+fmt.Sprint(latestOrder.ID), authHeader, nil, "GET")

			// unmarshal payment details
			var paymentDetailRes multiplePaymentRes
			paymentResErr := json.Unmarshal(paymentDetails, &paymentDetailRes)
			if paymentResErr != nil {
				fmt.Printf("error while unmarshalling payment details: %v", paymentResErr)
			}

			if len(paymentDetailRes.Data) == 0 {
				fmt.Printf("payment details not found for order id: %v", latestOrder.ID)
				// set latestOrderPaymentID as -1
				latestOrderPaymentID = new(int)
				*latestOrderPaymentID = -1
				// TODO: analyse and set the latestOrderPaymentStatus as well
			} else {
				latestOrderPaymentID = &paymentDetailRes.Data[0].ID
				latestOrderPaymentStatus = paymentDetailRes.Data[0].Status
			}

			// filter orderDetailsRes.Data to exclude cancelled orders
			var allPaidOrders []order
			for _, order := range orderDetailsRes.Data {
				if order.Status != "cancelled" {
					allPaidOrders = append(allPaidOrders, order)
				}
			}

			// Regular & Recurring paid order
			for i := len(allPaidOrders) - 1; i >= 0; i-- {

				// first order
				if i == len(allPaidOrders)-1 {
					if allPaidOrders[i].StartingDate.IsZero() {
						orderStartingDate = allPaidOrders[i].PaymentDate
					} else {
						orderStartingDate = allPaidOrders[i].StartingDate
					}
					previousStartingDate = orderStartingDate
					if allPaidOrders[i].Quantity != 0 {
						previousOrderQuantity = allPaidOrders[i].Quantity
					} else {
						previousOrderQuantity = 1
					}
				} else {
					// TODO: analyse and implement for scenarios where starting of the order is already available
					if allPaidOrders[i].PaymentDate.Before(previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)) {
						orderStartingDate = previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)
						previousStartingDate = orderStartingDate
						previousOrderQuantity = allPaidOrders[i].Quantity
					} else {
						orderStartingDate = allPaidOrders[i].PaymentDate
						previousStartingDate = orderStartingDate
						previousOrderQuantity = allPaidOrders[i].Quantity
					}

					if allPaidOrders[i].StartingDate.IsZero() {
						orderStartingDate = allPaidOrders[i].PaymentDate
					} else {
						orderStartingDate = allPaidOrders[i].StartingDate
					}
				}
				// update order via http call
				orderID := strconv.Itoa(allPaidOrders[i].ID)
				cancelOrder := getServerUrl() + "/pay/v2/order/" + orderID

				postBody, _ := json.Marshal(map[string]interface{}{
					"StartingDate": orderStartingDate,
				})

				buffPostBody := bytes.NewBuffer(postBody)

				_, resStatusCode := utils.HTTPCallAndGetBody(cancelOrder, authHeader, buffPostBody, "PATCH")

				if resStatusCode != 200 {
					fmt.Printf("error while updating order starting date: %v", resStatusCode)
					fmt.Printf("order id: %v", allPaidOrders[i].ID)
				}
			}
		}
	}

	if latestGrantCreatedAt != nil {
		if latestOrderPaymentDate != nil {
			if latestOrderPaymentDate.After(*latestGrantCreatedAt) {
				// update user grant cancelled_at to now
				_, grantUpdateErr := db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE id=$2`, time.Now(), *grantID)

				if grantUpdateErr != nil {
					return userMembershipRes{}, fmt.Errorf("problem updating grant: %w", grantUpdateErr)
				}
			} else {
				currentMembership = "helphaver"
			}
		} else {
			currentMembership = "helphaver"
		}
	}

	if currentMembership == "automatic" || currentMembership == "manual" {
		*membershipInsertData.Expiry = previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)
	} else if currentMembership == "helphaver" {
		// TODO: add integration of extra manual payment after grant is over
		userMonthsUsed, userMonthsUsedErr := db.getNumberOfGrantMonthsUsedByGrantID(ctx, *grantID)

		if userMonthsUsedErr != nil {
			fmt.Printf("error while getting user months used: %+v", userMonthsUsedErr)
		} else {
			userTotalMonthsLeft := *grantMonthsGranted - userMonthsUsed

			*membershipInsertData.Expiry = time.Now().AddDate(0, 0, 30*userTotalMonthsLeft)
		}
	}

	if membershipInsertData.Expiry == nil || membershipInsertData.Expiry.Before(time.Now()) {
		specialMembDetailUrl := getServerUrl() + "/pay/v2/special/" + email

		speMemDetails, statusCode := utils.HTTPCallAndGetBody(specialMembDetailUrl, authHeader, nil, "GET")

		if statusCode == 200 {
			// unmarshal payment details
			var speRes specialRes
			speResErr := json.Unmarshal(speMemDetails, &speRes)
			if speResErr != nil {
				fmt.Printf("error while unmarshalling special membership details: %+v", speResErr)
			} else {
				if speRes.Data.Email == email {
					userInSpecialTable = true
					currentMembership = "special"
					*membershipInsertData.Active = true
					*membershipInsertData.Type = "special"
					membershipInsertData.Expiry = nil
				}
			}
		} else if statusCode == 404 {
			userInSpecialTable = false
		} else {
			fmt.Printf("error while getting special membership details: %+v", statusCode)
		}

	}

	// check currentMembership is cancelled
	if allOrderCancelled || latestGrantCancelled {
		currentMembership = "cancelled"
	}

	if len(userGrant) == 0 && !userInSpecialTable && len(orderDetailsRes.Data) == 0 {
		currentMembership = "new"
	}

	membershipInsertData.Type = &currentMembership

	// check if user has an existing membership of the current month and year
	userMemb, userMembErr := db.getMultipleMembership(ctx, 0, 100, currentMonth, currentYear, userID)

	if userMembErr != nil {
		fmt.Printf("error while getting user membership: %+v", userMembErr)
	}

	// check if the membership is Active and Inactive
	if currentMembership != "special" {
		// check if expiry date is 60 days in past
		if membershipInsertData.Expiry != nil {
			if membershipInsertData.Expiry.Before(time.Now().AddDate(0, 0, -60)) {
				*membershipInsertData.Active = false
			} else {
				*membershipInsertData.Active = true
			}
		}
	} else {
		*membershipInsertData.Active = true
	}

	var (
		membershipID  int
		insertMembErr error
	)

	if len(userMemb) == 0 {
		// insert membership
		membershipID, insertMembErr = db.createMembership(ctx, membershipInsertData)

		if insertMembErr != nil {
			fmt.Printf("error while inserting membership: %+v", insertMembErr)
		}

		// remove all the prvious sub memberships entries as they'll be added again ( latest will be added )
		deleteAllMembershipSubTableErr := db.deleteAllMembershipSubTableByMembershipID(ctx, membershipID)

		if deleteAllMembershipSubTableErr != nil {
			fmt.Printf("error while deleting all membership sub table: %+v", deleteAllMembershipSubTableErr)
		}
	} else {
		// update membership
		updateMembErr := db.patchMembershipByID(ctx, membershipInsertData, *userMemb[0].ID)

		if updateMembErr != nil {
			fmt.Printf("error while updating membership: %+v", updateMembErr)
		}

	}

	if currentMembership == "automatic" {
		_, automaticErr := db.createAutomaticMembership(ctx, membershipAutomatic{
			MembershipID: &membershipID,
			OrderID:      &latestOrder.ID,
			PaymentID:    latestOrderPaymentID,
		})

		if automaticErr != nil {
			fmt.Printf("error while creating automatic membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", automaticErr)
		}
	} else if currentMembership == "manual" {
		_, manualErr := db.createManualMembership(ctx, membershipManual{
			MembershipID: &membershipID,
			OrderID:      &latestOrder.ID,
			PaymentID:    latestOrderPaymentID,
			Quantity:     &latestOrder.Quantity,
		})

		if manualErr != nil {
			fmt.Printf("error while creating manual membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", manualErr)
		}
	} else if currentMembership == "helphaver" {
		_, helphaverErr := db.createHelpHaverMembership(ctx, membershipHelpHaver{
			MembershipID: &membershipID,
			GrantID:      grantID,
			NbMonths:     grantMonthsGranted,
		})

		if helphaverErr != nil {
			fmt.Printf("error while creating helphaver membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", helphaverErr)
		}
	} else if currentMembership == "special" {
		_, specialErr := db.createSpecialMembership(ctx, membershipSpecial{
			MembershipID: &membershipID,
		})

		if specialErr != nil {
			fmt.Printf("error while creating special membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", specialErr)
		}
	}

	// add user notification
	var notificationSlugs []string
	if currentMembership == "automatic" || currentMembership == "manual" ||
		currentMembership == "cancelled" || currentMembership == "new" {

		if currentMembership == "automatic" && latestOrderPaymentStatus == "nosuccess" {
			notificationSlugs = append(notificationSlugs, "mb_problem_previous_payment")
		}

		if currentMembership == "manual" && time.Now().After(*membershipInsertData.Expiry) {
			notificationSlugs = append(notificationSlugs, "mb_expiration_notice")

		}

		if currentMembership == "cancelled" {
			notificationSlugs = append(notificationSlugs, "mb_cancelled")
		}

		if currentMembership == "new" {
			notificationSlugs = append(notificationSlugs, "mb_new")
		}
	}

	if len(userRequests) != 0 && latestRequest.ID != nil {
		if *latestRequest.Status == "REQUESTED" {
			notificationSlugs = append(notificationSlugs, "hh_request_received")
		}

		if latestGrantCreatedAt != nil {
			// status approved and latestGrantCreatedAt is less than a week old
			if *latestRequest.Status == "APPROVED" && latestGrantCreatedAt.After(time.Now().AddDate(0, 0, -7)) {
				notificationSlugs = append(notificationSlugs, "hh_request_approved")
			}

			// status rejected and latestRequest.UpdatedAt is less than a week old
			if *latestRequest.Status == "DENIED" && latestRequest.UpdatedAt.After(time.Now().AddDate(0, 0, -7)) {
				notificationSlugs = append(notificationSlugs, "hh_request_refused")
			}
		}

	}

	// add user notification if slug is not empty
	if len(notificationSlugs) != 0 {
		updateAllUserNotificationToInactiveErr := db.updateAllUserNotificationToInactive(ctx, userID)

		if updateAllUserNotificationToInactiveErr != nil {
			fmt.Printf("error while updating all user notification to inactive: %+v", updateAllUserNotificationToInactiveErr)
		}

		// loop through all the slugs and add notification
		for _, slug := range notificationSlugs {
			parentNotificationData, parentNotificationErr := db.getNotificationBySlug(ctx, slug)

			if parentNotificationErr != nil {
				fmt.Printf("error while getting parent notification: %+v", parentNotificationErr)
			} else {
				boolTrue := true
				userNotificationErr := db.createUserNotification(ctx, userNotification{
					UserID:         &userID,
					NotificationID: parentNotificationData.ID,
					Active:         &boolTrue,
					SeenAt:         nil,
				})

				if userNotificationErr != nil {
					fmt.Printf("error while creating notification: %+v", userNotificationErr)
				}
			}
		}
	}

	userMembershipResponse, userMembershipResponseErr := db.getMembershipByUserID(ctx, userID, authHeader)

	if userMembershipResponseErr != nil {
		fmt.Printf("error while getting membership by user id: %+v", userMembershipResponseErr)
	}

	return userMembershipResponse, nil
}

func (db *pgProfileDB) createMembership(ctx context.Context, req membership) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO membership (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) createManualMembership(ctx context.Context, req membershipManual) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareManualMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO membership_manual (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) createAutomaticMembership(ctx context.Context, req membershipAutomatic) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareAutomaticMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO membership_automatic (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) createSpecialMembership(ctx context.Context, req membershipSpecial) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareSpecialMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO membership_special (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) createHelpHaverMembership(ctx context.Context, req membershipHelpHaver) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareHelpHaverMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO membership_helphaver (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func prepareSpecialMembershipCreateQuery(req membershipSpecial) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.MembershipID != nil {
		createStrings = append(createStrings, "membership_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MembershipID)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareHelpHaverMembershipCreateQuery(req membershipHelpHaver) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.MembershipID != nil {
		createStrings = append(createStrings, "membership_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MembershipID)
	}
	if req.GrantID != nil {
		createStrings = append(createStrings, "grant_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.GrantID)
	}
	if req.NbMonths != nil {
		createStrings = append(createStrings, "nb_months")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.NbMonths)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareAutomaticMembershipCreateQuery(req membershipAutomatic) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.OrderID != nil {
		createStrings = append(createStrings, "order_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.OrderID)
	}
	if req.PaymentID != nil {
		createStrings = append(createStrings, "payment_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.PaymentID)
	}
	if req.MembershipID != nil {
		createStrings = append(createStrings, "membership_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MembershipID)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareManualMembershipCreateQuery(req membershipManual) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.OrderID != nil {
		createStrings = append(createStrings, "order_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.OrderID)
	}
	if req.PaymentID != nil {
		createStrings = append(createStrings, "payment_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.PaymentID)
	}
	if req.MembershipID != nil {
		createStrings = append(createStrings, "membership_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MembershipID)
	}
	if req.Quantity != nil {
		createStrings = append(createStrings, "quantity")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Quantity)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareMembershipCreateQuery(req membership) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Active != nil {
		createStrings = append(createStrings, "active")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Active)
	}
	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}
	if req.Expiry != nil {
		createStrings = append(createStrings, "expiry")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Expiry)
	}
	if req.UserID != nil {
		createStrings = append(createStrings, "user_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.UserID)
	}
	if req.Month != nil {
		createStrings = append(createStrings, "month")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Month)
	}
	if req.Year != nil {
		createStrings = append(createStrings, "year")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Year)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func (db *pgProfileDB) getMembershipByID(ctx context.Context, id int) (membershipRes, error) {
	var membership membershipRes

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		active,
		user_id,
		type,
		month,
		year,
		expiry,
		created_at,
		updated_at,
		deleted_at
		FROM membership 
		WHERE id = $1`, id).Scan(
		&membership.ID,
		&membership.Active,
		&membership.UserID,
		&membership.Type,
		&membership.Month,
		&membership.Year,
		&membership.Expiry,
		&membership.CreatedAt,
		&membership.UpdatedAt,
		&membership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return membershipRes{}, errNotFound
		}
		return membershipRes{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return membership, nil
}

func (db *pgProfileDB) getMembershipByUserID(ctx context.Context, userID string, authHeader string) (userMembershipRes, error) {
	var membership userMembershipRes

	// fetch current month in int
	currentMonth := time.Now().Month()
	// fetch current year
	currentYear := time.Now().Year()

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		active,
		user_id,
		type,
		month,
		year,
		expiry,
		created_at,
		updated_at,
		deleted_at
		FROM membership 
		WHERE user_id = $1 AND month = $2 AND year = $3`, userID, currentMonth, currentYear).Scan(
		&membership.ID,
		&membership.Active,
		&membership.UserID,
		&membership.Type,
		&membership.Month,
		&membership.Year,
		&membership.Expiry,
		&membership.CreatedAt,
		&membership.UpdatedAt,
		&membership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return userMembershipRes{}, errNotFound
		}
		return userMembershipRes{}, fmt.Errorf("error while getting membership: %w", err)
	}

	if *membership.Type == "automatic" {
		autoMembership, err := db.getAutomaticMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return userMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}
		membership.Details.Automatic.OrderID = autoMembership.OrderID
		membership.Details.Automatic.PaymentID = autoMembership.PaymentID

		// Get payment details from order service
		userPaymentDetails := getServerUrl() + "/pay/v2/payment/" + fmt.Sprint(autoMembership.PaymentID)

		orderDetails, _ := utils.HTTPCallAndGetBody(userPaymentDetails, authHeader, nil, "GET")

		// unmarshal payment details
		var paymentDetailRes paymentRes
		paymentResErr := json.Unmarshal(orderDetails, &paymentDetailRes)
		if paymentResErr != nil {
			return userMembershipRes{}, fmt.Errorf("error while unmarshalling payment details: %w", paymentResErr)
		}

		// Add payment details to membership
		membership.Details.Payment.Amount = &paymentDetailRes.Data.Amount
		membership.Details.Payment.Currency = &paymentDetailRes.Data.DebitCurrency
		membership.Details.Payment.Status = &paymentDetailRes.Data.PaymentStatus
		membership.Details.Payment.Date = &paymentDetailRes.Data.CreatedAt
		membership.Details.Payment.PaymentMethod = &paymentDetailRes.Data.CCNumber

	} else if *membership.Type == "helphaver" {
		helphaverMembership, err := db.getHelphaverMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return userMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}
		membership.Details.HelpHaver.CreatedAt = helphaverMembership.CreatedAt
		membership.Details.HelpHaver.NbMonths = helphaverMembership.NbMonths
	} else if *membership.Type == "manual" {
		manualMembership, err := db.getManualMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return userMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}

		manualUserPaymentDetails := getServerUrl() + "/pay/v2/payment/" + fmt.Sprint(manualMembership.PaymentID)

		paymentDetails, _ := utils.HTTPCallAndGetBody(manualUserPaymentDetails, authHeader, nil, "GET")

		// unmarshal payment details
		var paymentDetailRes paymentRes
		paymentResErr := json.Unmarshal(paymentDetails, &paymentDetailRes)
		if paymentResErr != nil {
			return userMembershipRes{}, fmt.Errorf("error while unmarshalling payment details: %w", paymentResErr)
		}

		// Add payment details to membership
		membership.Details.Payment.Amount = &paymentDetailRes.Data.Amount
		membership.Details.Payment.Currency = &paymentDetailRes.Data.DebitCurrency
		membership.Details.Payment.Status = &paymentDetailRes.Data.PaymentStatus
		membership.Details.Payment.Date = &paymentDetailRes.Data.CreatedAt
		membership.Details.Payment.PaymentMethod = &paymentDetailRes.Data.CCNumber

	} else if *membership.Type == "special" {

		var userEmail string

		if err := db.QueryRow(ctx, `SELECT primary_email FROM users WHERE user_id=$1`, userID).Scan(&userEmail); err != nil {
			return userMembershipRes{}, fmt.Errorf("error while getting user email: %w", err)
		}

		specialMembDetailUrl := getServerUrl() + "/pay/v2/special/" + userEmail

		speMemDetails, _ := utils.HTTPCallAndGetBody(specialMembDetailUrl, authHeader, nil, "GET")

		// unmarshal payment details
		var speRes specialRes
		speResErr := json.Unmarshal(speMemDetails, &speRes)
		if speResErr != nil {
			return userMembershipRes{}, fmt.Errorf("error while unmarshalling special membership details: %w", speResErr)
		}

		// Add special membership details to membership
		membership.Details.Special.ApprovedBy = &speRes.Data.Category
		membership.Details.Special.Type = &speRes.Data.SubCategory

	}

	// fetch active user notification
	userActiveNotification, userNotiErr := db.getActiveUserNotificationByUserID(ctx, userID)

	if userNotiErr != nil {
		return userMembershipRes{}, fmt.Errorf("error while getting active user notification: %w", userNotiErr)
	}

	var userNotiSlug []userMembershipNotification
	// loop and add slug to membership
	for _, noti := range userActiveNotification {
		var userNoti userMembershipNotification
		userNoti.Slug = noti.Slug
		userNotiSlug = append(userNotiSlug, userNoti)
	}

	membership.Notifications = userNotiSlug

	return membership, nil
}

func (db *pgProfileDB) getAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (membershipAutomatic, error) {
	var autoMembership membershipAutomatic

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		order_id,
		payment_id,
		membership_id,
		created_at,
		updated_at,
		deleted_at 
		FROM membership_automatic 
		WHERE membership_id = $1`, membershipID).Scan(
		&autoMembership.ID,
		&autoMembership.OrderID,
		&autoMembership.PaymentID,
		&autoMembership.MembershipID,
		&autoMembership.CreatedAt,
		&autoMembership.UpdatedAt,
		&autoMembership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return membershipAutomatic{}, errNotFound
		}
		return membershipAutomatic{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return autoMembership, nil
}

func (db *pgProfileDB) getSpecialMembershipByMembershipID(ctx context.Context, membershipID int) (membershipSpecial, error) {
	var specialMembership membershipSpecial

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		membership_id,
		created_at,
		updated_at,
		deleted_at 
		FROM membership_special 
		WHERE membership_id = $1`, membershipID).Scan(
		&specialMembership.ID,
		&specialMembership.MembershipID,
		&specialMembership.CreatedAt,
		&specialMembership.UpdatedAt,
		&specialMembership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return membershipSpecial{}, errNotFound
		}
		return membershipSpecial{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return specialMembership, nil
}

func (db *pgProfileDB) getManualMembershipByMembershipID(ctx context.Context, membershipID int) (membershipManual, error) {
	var manualMembership membershipManual

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		order_id,
		payment_id,
		membership_id,
		quantity,
		created_at,
		updated_at,
		deleted_at 
		FROM membership_manual 
		WHERE membership_id = $1`, membershipID).Scan(
		&manualMembership.ID,
		&manualMembership.OrderID,
		&manualMembership.PaymentID,
		&manualMembership.MembershipID,
		&manualMembership.Quantity,
		&manualMembership.CreatedAt,
		&manualMembership.UpdatedAt,
		&manualMembership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return membershipManual{}, errNotFound
		}
		return membershipManual{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return manualMembership, nil
}

func (db *pgProfileDB) getHelphaverMembershipByMembershipID(ctx context.Context, membershipID int) (membershipHelpHaver, error) {
	var helphaverMembership membershipHelpHaver

	if err := db.QueryRow(ctx, `
		SELECT 
		membership_helphaver.id,
		grant_id,
		membership_id,
		"grant".nb_months,
		created_at,
		updated_at,
		deleted_at 
		from membership_helphaver LEFT JOIN "grant" ON membership_helphaver.grant_id = "grant".id 
		WHERE membership_id = $1`, membershipID).Scan(
		&helphaverMembership.ID,
		&helphaverMembership.GrantID,
		&helphaverMembership.MembershipID,
		&helphaverMembership.NbMonths,
		&helphaverMembership.CreatedAt,
		&helphaverMembership.UpdatedAt,
		&helphaverMembership.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return membershipHelpHaver{}, errNotFound
		}
		return membershipHelpHaver{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return helphaverMembership, nil
}

func (db *pgProfileDB) patchMembershipByID(ctx context.Context, membership membership, id int) error {

	toUpdate, toUpdateArgs := prepareMembershipUpdateQuery(membership)

	if len(toUpdateArgs) != 0 {
		updateRes, err := db.Exec(ctx, fmt.Sprintf(`UPDATE membership SET %s WHERE id=%d`, toUpdate, id),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating membership: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) cancelMembership(ctx context.Context, membBody emailKeycloakAndUserIDBody, authHeader string) (int, int, int, error) {

	var email string
	var user_id string

	if membBody.UserID != nil && *membBody.UserID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email FROM users WHERE user_id=$1`, *membBody.UserID).Scan(&email); err != nil {
			return 0, 0, 0, fmt.Errorf("error while getting user email: %w", err)
		}
		user_id = *membBody.UserID
	} else if membBody.KeycloakID != nil && *membBody.KeycloakID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, user_id FROM users WHERE keycloak_id=$1`, *membBody.KeycloakID).Scan(&email, &user_id); err != nil {
			return 0, 0, 0, fmt.Errorf("error while getting user email: %w", err)
		}
	} else {
		if err := db.QueryRow(ctx, `SELECT user_id FROM users WHERE primary_email=$1`, *membBody.Email).Scan(&user_id); err != nil {
			return 0, 0, 0, fmt.Errorf("error while getting user id: %w", err)
		}
		email = *membBody.Email
	}

	tx, txErr := db.Begin(ctx)

	if txErr != nil {
		return 0, 0, 0, txErr
	}

	defer func() { _ = tx.Rollback(ctx) }()

	// implemennt to only update orders where status is not equal to cancelled
	userOrderDetails := getServerUrl() + "/pay/v2/orders?email=" + email + "&product-type=globalmembership&limit=100"

	orderDetails, _ := utils.HTTPCallAndGetBody(userOrderDetails, authHeader, nil, "GET")

	// un marshal order details
	var orderDetailsRes orderRes
	err := json.Unmarshal(orderDetails, &orderDetailsRes)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error while unmarshalling order details: %w", err)
	}

	// loop over order details and cancel the order
	for _, order := range orderDetailsRes.Data {
		// id int to string
		orderID := fmt.Sprintf("%d", order.ID)
		cancelOrder := getServerUrl() + "/pay/v2/order/" + orderID
		postBody, _ := json.Marshal(map[string]interface{}{
			"Status": "cancelled",
		})

		buffPostBody := bytes.NewBuffer(postBody)

		cancelOrderRes, _ := utils.HTTPCallAndGetBody(cancelOrder, authHeader, buffPostBody, "PATCH")

		// un marshal cancel order details
		var cancelOrderResRes orderRes
		err := json.Unmarshal(cancelOrderRes, &cancelOrderResRes)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error while unmarshalling cancel order details: %w", err)
		}
	}

	// update user grant cancelled_at to now

	//double check
	grantUpdateRes, grantUpdateErr := db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE user_id=$2 AND type='membership' AND cancelled_at IS NULL`, time.Now(), user_id)

	if grantUpdateErr != nil {
		return 0, 0, 0, fmt.Errorf("error while updating grant: %w", grantUpdateErr)
	}

	numberOfRowsUpdated := grantUpdateRes.RowsAffected()

	specialTableDelete := getServerUrl() + "/pay/v2/special/" + email

	specialTableDelRes, _ := utils.HTTPCallAndGetBody(specialTableDelete, authHeader, nil, "DELETE")

	// un marshal special table delete details
	var specialTableDelResRes orderDeleteRes

	err = json.Unmarshal(specialTableDelRes, &specialTableDelResRes)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("error while unmarshalling special table delete details: %w", err)
	}

	return len(orderDetailsRes.Data), int(numberOfRowsUpdated), specialTableDelResRes.Data, tx.Commit(ctx)

}

func (db *pgProfileDB) softDeleteMembershipByID(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `UPDATE membership SET deleted_at=$1 WHERE id=$2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("problem soft deleting membership: %w", err)
	}
	return nil
}

func (db *pgProfileDB) deleteAllMembershipSubTableByMembershipID(ctx context.Context, membershipID int) error {
	_, err := db.Exec(ctx, `
	BEGIN;

	DELETE FROM membership_manual WHERE membership_id=$1;
	DELETE FROM membership_special WHERE membership_id=$1;
	DELETE FROM membership_automatic WHERE membership_id=$1;
	DELETE FROM membership_helphaver WHERE membership_id=$1;
	
	COMMIT;
	`, membershipID)
	if err != nil {
		return fmt.Errorf("problem soft deleting membership: %w", err)
	}
	return nil
}

func (db *pgProfileDB) getMultipleMembership(ctx context.Context, intSkip int, intLimit int, month int, year int, userID string) ([]membershipRes, error) {
	memberships := []membershipRes{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereMembershipQuery(month, year, userID)

	rows, err := db.Query(ctx, `
		SELECT 
		id,
		active,
		user_id,
		type,
		month,
		year,
		expiry,
		created_at,
		updated_at,
		deleted_at
		FROM membership `+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []membershipRes{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r membershipRes
		if err := rows.Scan(
			&r.ID,
			&r.Active,
			&r.UserID,
			&r.Type,
			&r.Month,
			&r.Year,
			&r.Expiry,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []membershipRes{}, err
		}

		memberships = append(memberships, r)
	}

	return memberships, nil
}

func buildAndGetWhereMembershipQuery(month int, year int, userID string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	if month != 0 {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND month='%d'", month))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" month='%d'", month))
		}
	}

	if year != 0 {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND year='%d'", year))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" year='%d'", year))
		}
	}

	if userID != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND user_id='%s'", userID))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" user_id='%s'", userID))
		}
	}

	orderBy.WriteString(fmt.Sprintf(" ORDER BY updated_at %s", "desc"))

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}

func prepareMembershipUpdateQuery(req membership) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Active != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("active=$%d", len(updateStrings)+1))
		args = append(args, *req.Active)
	}

	if req.Expiry != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("expiry=$%d", len(updateStrings)+1))
		args = append(args, *req.Expiry)
	}

	if req.Month != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("month=$%d", len(updateStrings)+1))
		args = append(args, *req.Month)
	}

	if req.Year != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("year=$%d", len(updateStrings)+1))
		args = append(args, *req.Year)
	}

	if req.UserID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("user_id=$%d", len(updateStrings)+1))
		args = append(args, *req.UserID)
	}

	if req.Type != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("type=$%d", len(updateStrings)+1))
		args = append(args, *req.Type)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
