package repo

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
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

type membershipInterface interface {
	GetMembershipByID(ctx context.Context, id int) (Membership, error)
	GetMembershipByUserID(ctx context.Context, userID string, authHeader string) (UserMembershipRes, error)
	GetMembershipByKCID(ctx context.Context, kcID string, authHeader string) (UserMembershipRes, error)
	GetMultipleMembership(ctx context.Context, intSkip int, intLimit int, month int, year int, userID string) ([]Membership, error)
	PatchMembershipByID(ctx context.Context, membership Membership, id int) (int, error)
	SoftDeleteMembershipByID(ctx context.Context, id int) error
	CancelMembership(ctx context.Context, body EmailKeycloakAndUserIDBody, authHeader string) (int, int, int, error)
	GetAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipAutomatic, error)
	EvaluateMembershipByUserID(ctx context.Context, evalbody EmailKeycloakAndUserIDBody, authHeader string) (UserMembershipRes, error)
}

type Membership struct {
	ID        *int       `json:"id" db:"id"`
	Active    *bool      `json:"active"`
	UserID    *uuid.UUID `json:"user_id"`
	Type      *string    `json:"type"`
	Month     *int       `json:"month"`
	Year      *int       `json:"year"`
	Expiry    *time.Time `json:"expiry"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

type MembershipAutomatic struct {
	ID           *int       `json:"id"`
	OrderID      *int       `json:"order_id"`
	PaymentID    *int       `json:"payment_id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type MembershipSpecial struct {
	ID           *int       `json:"id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type MembershipManual struct {
	ID           *int       `json:"id"`
	OrderID      *int       `json:"order_id"`
	PaymentID    *int       `json:"payment_id"`
	MembershipID *int       `json:"membership_id"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	Quantity     *int       `json:"quantity"`
}

type MembershipHelpHaver struct {
	ID           *int       `json:"id"`
	GrantID      *int       `json:"grant_id"`
	MembershipID *int       `json:"membership_id"`
	NbMonths     *int       `json:"nb_months,omitempty"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type UserMembershipRes struct {
	Membership
	Notifications []UserMembershipNotification `json:"notifications,omitempty"`
	Details       struct {
		Payment struct {
			Date          *time.Time `json:"date,omitempty"`
			Amount        *int       `json:"amount,omitempty"`
			Currency      *string    `json:"currency,omitempty"`
			PaymentMethod *string    `json:"payment_method,omitempty"`
			Status        *string    `json:"status,omitempty"`
		} `json:"payment,omitempty"`
		Automatic struct {
			OrderID   *int `json:"order_id,omitempty"`
			PaymentID *int `json:"payment_id,omitempty"`
		} `json:"automatic"`
		Special struct {
			ApprovedBy *string `json:"approved_by,omitempty"`
			Type       *string `json:"type,omitempty"`
		} `json:"special,omitempty"`
		HelpHaver struct {
			CreatedAt *time.Time `json:"created_at,omitempty"`
			NbMonths  *int       `json:"nb_months,omitempty"`
		} `json:"help_haver,omitempty"`
	} `json:"details,omitempty"`
}

type UserMembershipNotification struct {
	Slug    *string                 `json:"slug"`
	Content *map[string]interface{} `json:"content,omitempty"`
}

type EmailKeycloakAndUserIDBody struct {
	Email      *string `json:"email"`
	KeycloakID *string `json:"keycloak_id"`
	UserID     *string `json:"user_id"`
}

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
func (db *ProfileDB) EvaluateMembershipByUserID(ctx context.Context, evalBody EmailKeycloakAndUserIDBody, authHeader string) (UserMembershipRes, error) {

	var email string
	var userID string
	var userKeycloakID string

	// fetch email, userID and userKeycloakID from the database if not available in the request body
	if evalBody.UserID != nil && *evalBody.UserID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, keycloak_id FROM users WHERE user_id=$1`, *evalBody.UserID).Scan(&email, &userKeycloakID); err != nil {
			return UserMembershipRes{}, fmt.Errorf("problem getting email from user_id: %w", err)
		}
		userID = *evalBody.UserID
	} else if evalBody.KeycloakID != nil && *evalBody.KeycloakID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, user_id FROM users WHERE keycloak_id=$1`, *evalBody.KeycloakID).Scan(&email, &userID); err != nil {
			return UserMembershipRes{}, fmt.Errorf("problem getting email from keycloak_id: %w", err)
		}
		userKeycloakID = *evalBody.KeycloakID
	} else {
		if err := db.QueryRow(ctx, `SELECT user_id, keycloak_id FROM users WHERE primary_email=$1`, *evalBody.Email).Scan(&userID, &userKeycloakID); err != nil {
			return UserMembershipRes{}, fmt.Errorf("problem getting user_id from email: %w", err)
		}
		email = *evalBody.Email
	}

	var uuidErr error

	var membershipInsertData Membership
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
	var latestRequest RequestResponse

	const LIMIT = 200
	// limit to string

	var orderStartingDate time.Time
	var previousStartingDate time.Time
	var previousOrderQuantity int
	currentMonth := int(time.Now().Month())
	currentYear := time.Now().Year()
	// default value of active is false
	membershipInsertData.Active = utils.PointerBool(false)
	membershipInsertData.Month = &currentMonth
	membershipInsertData.Year = &currentYear

	var userUUID uuid.UUID
	userUUID, uuidErr = uuid.FromString(userID)
	membershipInsertData.UserID = &userUUID

	if uuidErr != nil {
		return UserMembershipRes{}, fmt.Errorf("problem converting userID to uuid: %w", uuidErr)
	}

	// if not email exit
	if email == "" {
		return UserMembershipRes{}, fmt.Errorf("no email found")
	}

	// check if user has any active request
	userRequests, userRequestsErr := db.GetMultipleRequest(ctx, 0, LIMIT, userKeycloakID, "", "", "hhmembership", "desc")

	if userRequestsErr != nil {
		return UserMembershipRes{}, fmt.Errorf("problem getting User requests: %w", userRequestsErr)
	}

	if len(userRequests) != 0 {
		// latest request of the user
		latestRequest = userRequests[0]
	}

	// fetch all the order of the user ( limit 200 as of now )
	userOrderDetails := common.GetOrdersServiceUrl() + "/v2/orders?email=" + email +
		"&product-type=globalmembership&limit=" + strconv.Itoa(LIMIT) +
		"&evaluate-membership=true&o-payment-date=desc"
	orderDetails, _ := utils.HTTPCallAndGetBody(userOrderDetails, authHeader, nil, "GET")

	// unmarshal order details
	var orderDetailsRes orderRes
	err := json.Unmarshal(orderDetails, &orderDetailsRes)
	if err != nil {
		fmt.Println("problem unmarshalling order details: %w", err)
	}

	// fetch user grant if any
	userGrant, userGrantErr := db.GetMultipleGrant(ctx, 0, LIMIT, nil, userID, "hhmembership", "desc")

	if userGrantErr != nil {
		if !errors.Is(userGrantErr, common.ErrUserNotFound) {
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
			if errors.Is(grantMembErr, common.ErrNotFound) {
				fmt.Println("no grant membership found for grant id: ", *grantID)
			} else {
				return UserMembershipRes{}, fmt.Errorf("problem getting grant membership: %w", grantMembErr)
			}
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
			paymentDetails, _ := utils.HTTPCallAndGetBody(common.GetOrdersServiceUrl()+"/v2/payments?o-created-at=desc&order-id="+fmt.Sprint(latestOrder.ID), authHeader, nil, "GET")

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
				cancelOrder := common.GetOrdersServiceUrl() + "/v2/order/" + orderID
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
					return UserMembershipRes{}, fmt.Errorf("problem updating grant: %w", grantUpdateErr)
				}
			} else {
				currentMembership = "helphaver"
			}
		} else {
			currentMembership = "helphaver"
		}
	}

	if currentMembership == "automatic" || currentMembership == "manual" {
		var newDate = previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)
		membershipInsertData.Expiry = &newDate

		// check if final orderStartingDate is more than 90 days in past from now
		if orderStartingDate.Before(time.Now().AddDate(0, 0, -90)) {
			membershipInsertData.Active = utils.PointerBool(false)
		}

	} else if currentMembership == "helphaver" {
		// TODO: add integration of extra manual payment after grant is over
		userMonthsUsed, userMonthsUsedErr := db.getNumberOfGrantMonthsUsedByGrantID(ctx, *grantID)

		if userMonthsUsedErr != nil {
			fmt.Printf("error while getting user months used: %+v", userMonthsUsedErr)
		} else {
			userTotalMonthsLeft := *grantMonthsGranted - userMonthsUsed

			var newExpiryDate = time.Now().AddDate(0, 0, 30*userTotalMonthsLeft)
			membershipInsertData.Expiry = &newExpiryDate
		}
	}

	if membershipInsertData.Expiry == nil || membershipInsertData.Expiry.Before(time.Now()) {
		specialMembDetailUrl := common.GetOrdersServiceUrl() + "/v2/special/" + email
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
	userMemb, userMembErr := db.GetMultipleMembership(ctx, 0, 100, currentMonth, currentYear, userID)

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
		updateMembErr error
	)

	if len(userMemb) == 0 {
		// insert membership
		membershipID, insertMembErr = db.createMembership(ctx, membershipInsertData)

		if insertMembErr != nil {
			fmt.Printf("error while inserting membership: %+v", insertMembErr)
		}

		// remove all the prvious sub memberships entries as they'll be added again ( latest will be added )
		// #check with yasha
		deleteAllMembershipSubTableErr := db.deleteAllMembershipSubTableByMembershipID(ctx, membershipID)

		if deleteAllMembershipSubTableErr != nil {
			fmt.Printf("error while deleting all membership sub table: %+v", deleteAllMembershipSubTableErr)
		}
	} else {
		// update membership
		membershipID, updateMembErr = db.PatchMembershipByID(ctx, membershipInsertData, *userMemb[0].ID)

		if updateMembErr != nil {
			fmt.Printf("error while updating membership: %+v", updateMembErr)
		}

	}

	if currentMembership == "automatic" {
		_, automaticErr := db.createAutomaticMembership(ctx, MembershipAutomatic{
			MembershipID: &membershipID,
			OrderID:      &latestOrder.ID,
			PaymentID:    latestOrderPaymentID,
		})

		if automaticErr != nil {
			fmt.Printf("error while creating automatic membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", automaticErr)
		}
	} else if currentMembership == "manual" {
		_, manualErr := db.createManualMembership(ctx, MembershipManual{
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
		_, helphaverErr := db.createHelpHaverMembership(ctx, MembershipHelpHaver{
			MembershipID: &membershipID,
			GrantID:      grantID,
		})

		if helphaverErr != nil {
			fmt.Printf("error while creating helphaver membership with membership id: %v", membershipID)
			fmt.Printf("error: %v", helphaverErr)
		}
	} else if currentMembership == "special" {
		_, specialErr := db.createSpecialMembership(ctx, MembershipSpecial{
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
			// check if final orderStartingDate is more than 90 days in past from now
			if orderStartingDate.Before(time.Now().AddDate(0, 0, -90)) {
				notificationSlugs = append(notificationSlugs, "mb_has_expired_notice")
			}
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
				userNotificationErr := db.CreateUserNotification(ctx, UserNotification{
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

	userMembershipResponse, userMembershipResponseErr := db.GetMembershipByUserID(ctx, userID, authHeader)

	if userMembershipResponseErr != nil {
		fmt.Printf("error while getting membership by user id: %+v", userMembershipResponseErr)
	}

	return userMembershipResponse, nil
}

func (db *ProfileDB) createMembership(ctx context.Context, req Membership) (int, error) {

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

func (db *ProfileDB) createManualMembership(ctx context.Context, req MembershipManual) (int, error) {

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

func (db *ProfileDB) createAutomaticMembership(ctx context.Context, req MembershipAutomatic) (int, error) {

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

func (db *ProfileDB) createSpecialMembership(ctx context.Context, req MembershipSpecial) (int, error) {

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

func (db *ProfileDB) createHelpHaverMembership(ctx context.Context, req MembershipHelpHaver) (int, error) {

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

func prepareSpecialMembershipCreateQuery(req MembershipSpecial) (string, string, []interface{}) {
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

func prepareHelpHaverMembershipCreateQuery(req MembershipHelpHaver) (string, string, []interface{}) {
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

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareAutomaticMembershipCreateQuery(req MembershipAutomatic) (string, string, []interface{}) {
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

func prepareManualMembershipCreateQuery(req MembershipManual) (string, string, []interface{}) {
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

func prepareMembershipCreateQuery(req Membership) (string, string, []interface{}) {
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

func (db *ProfileDB) GetMembershipByID(ctx context.Context, id int) (Membership, error) {
	var membership Membership

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
			return Membership{}, common.ErrNotFound
		}
		return Membership{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return membership, nil
}

func (db *ProfileDB) GetMembershipByKCID(ctx context.Context, kcID string, authHeader string) (UserMembershipRes, error) {
	var userID uuid.UUID
	if err := db.QueryRow(ctx, "SELECT user_id FROM users WHERE keycloak_id=$1", kcID).
		Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserMembershipRes{}, common.ErrNotFound
		}
		return UserMembershipRes{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return db.GetMembershipByUserID(ctx, userID.String(), authHeader)
}

func (db *ProfileDB) GetMembershipByUserID(ctx context.Context, userID string, authHeader string) (UserMembershipRes, error) {
	var membership UserMembershipRes

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
			return UserMembershipRes{}, common.ErrNotFound
		}
		return UserMembershipRes{}, fmt.Errorf("error while getting membership: %w", err)
	}

	if *membership.Type == "automatic" {
		autoMembership, err := db.GetAutomaticMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}
		membership.Details.Automatic.OrderID = autoMembership.OrderID
		membership.Details.Automatic.PaymentID = autoMembership.PaymentID

		// Get payment details from order service
		userPaymentDetails := common.GetOrdersServiceUrl() + "/v2/payment/" + fmt.Sprint(*autoMembership.PaymentID)
		orderDetails, _ := utils.HTTPCallAndGetBody(userPaymentDetails, authHeader, nil, "GET")

		// unmarshal payment details
		var paymentDetailRes paymentRes
		paymentResErr := json.Unmarshal(orderDetails, &paymentDetailRes)
		if paymentResErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error while unmarshalling payment details: %w", paymentResErr)
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
			return UserMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}
		membership.Details.HelpHaver.CreatedAt = helphaverMembership.CreatedAt
		membership.Details.HelpHaver.NbMonths = helphaverMembership.NbMonths
	} else if *membership.Type == "manual" {
		manualMembership, err := db.getManualMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}

		manualUserPaymentDetails := common.GetOrdersServiceUrl() + "/v2/payment/" + fmt.Sprint(*manualMembership.PaymentID)
		paymentDetails, _ := utils.HTTPCallAndGetBody(manualUserPaymentDetails, authHeader, nil, "GET")

		// unmarshal payment details
		var paymentDetailRes paymentRes
		paymentResErr := json.Unmarshal(paymentDetails, &paymentDetailRes)
		if paymentResErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error while unmarshalling payment details: %w", paymentResErr)
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
			return UserMembershipRes{}, fmt.Errorf("error while getting User email: %w", err)
		}

		specialMembDetailUrl := common.GetOrdersServiceUrl() + "/v2/special/" + userEmail
		speMemDetails, _ := utils.HTTPCallAndGetBody(specialMembDetailUrl, authHeader, nil, "GET")

		// unmarshal payment details
		var speRes specialRes
		speResErr := json.Unmarshal(speMemDetails, &speRes)
		if speResErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error while unmarshalling special membership details: %w", speResErr)
		}

		// Add special membership details to membership
		membership.Details.Special.ApprovedBy = &speRes.Data.Category
		membership.Details.Special.Type = &speRes.Data.SubCategory

	}

	// fetch active user notification
	userActiveNotification, userNotiErr := db.getActiveUserNotificationByUserID(ctx, userID)

	if userNotiErr != nil {
		return UserMembershipRes{}, fmt.Errorf("error while getting active user notification: %w", userNotiErr)
	}

	var userNotiSlug []UserMembershipNotification
	// loop and add slug to membership
	for _, noti := range userActiveNotification {
		var userNoti UserMembershipNotification
		userNoti.Slug = noti.Slug
		userNoti.Content = noti.Content
		userNotiSlug = append(userNotiSlug, userNoti)
	}

	membership.Notifications = userNotiSlug

	return membership, nil
}

func (db *ProfileDB) GetAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipAutomatic, error) {
	var autoMembership MembershipAutomatic

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
			return MembershipAutomatic{}, common.ErrNotFound
		}
		return MembershipAutomatic{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return autoMembership, nil
}

func (db *ProfileDB) getSpecialMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipSpecial, error) {
	var specialMembership MembershipSpecial

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
			return MembershipSpecial{}, common.ErrNotFound
		}
		return MembershipSpecial{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return specialMembership, nil
}

func (db *ProfileDB) getManualMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipManual, error) {
	var manualMembership MembershipManual

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
			return MembershipManual{}, common.ErrNotFound
		}
		return MembershipManual{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return manualMembership, nil
}

func (db *ProfileDB) getHelphaverMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipHelpHaver, error) {
	var helphaverMembership MembershipHelpHaver

	if err := db.QueryRow(ctx, `
		SELECT 
		membership_helphaver.id,
		membership_helphaver.grant_id,
		membership_id,
		grant_membership.nb_months,
		membership_helphaver.created_at,
		membership_helphaver.updated_at,
		membership_helphaver.deleted_at 
		from membership_helphaver LEFT JOIN grant_membership ON membership_helphaver.grant_id = grant_membership.grant_id 
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
			return MembershipHelpHaver{}, common.ErrNotFound
		}
		return MembershipHelpHaver{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return helphaverMembership, nil
}

func (db *ProfileDB) PatchMembershipByID(ctx context.Context, membership Membership, id int) (int, error) {

	toUpdate, toUpdateArgs := prepareMembershipUpdateQuery(membership)

	var membershipID int

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE membership SET %s WHERE id=%d returning id`, toUpdate, id),
			toUpdateArgs...).
			Scan(&membershipID); err != nil {
			if err == pgx.ErrNoRows {
				return 0, common.ErrNotFound
			}
			return 0, fmt.Errorf("problem updating membership: %w", err)
		}

		return membershipID, nil
	} else {
		return 0, fmt.Errorf("no fields to update")
	}
}

func (db *ProfileDB) CancelMembership(ctx context.Context, membBody EmailKeycloakAndUserIDBody, authHeader string) (int, int, int, error) {

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

	// implement to only update orders where status is not equal to cancelled
	userOrderDetails := common.GetOrdersServiceUrl() + "/v2/orders?email=" + email + "&product-type=globalmembership&limit=100"
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
		cancelOrder := common.GetOrdersServiceUrl() + "/v2/order/" + orderID
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

	specialTableDelete := common.GetOrdersServiceUrl() + "/v2/special/" + email
	specialTableDelRes, _ := utils.HTTPCallAndGetBody(specialTableDelete, authHeader, nil, "DELETE")

	// un marshal special table delete details
	var specialTableDelResRes orderDeleteRes

	err = json.Unmarshal(specialTableDelRes, &specialTableDelResRes)

	if err != nil {
		return 0, 0, 0, fmt.Errorf("error while unmarshalling special table delete details: %w", err)
	}

	return len(orderDetailsRes.Data), int(numberOfRowsUpdated), specialTableDelResRes.Data, tx.Commit(ctx)

}

func (db *ProfileDB) SoftDeleteMembershipByID(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `UPDATE membership SET deleted_at=$1 WHERE id=$2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("problem soft deleting membership: %w", err)
	}
	return nil
}

func (db *ProfileDB) deleteAllMembershipSubTableByMembershipID(ctx context.Context, membershipID int) error {

	// start transaction
	tx, txErr := db.Begin(ctx)
	if txErr != nil {
		return fmt.Errorf("problem starting transaction: %w", txErr)
	}

	// run delete on all sub tables
	_, membership_manualErr := tx.Exec(ctx, `DELETE FROM membership_manual WHERE membership_id=$1`, membershipID)
	_, membership_specialErr := tx.Exec(ctx, `DELETE FROM membership_special WHERE membership_id=$1`, membershipID)
	_, membership_automaticErr := tx.Exec(ctx, `DELETE FROM membership_automatic WHERE membership_id=$1`, membershipID)
	_, membership_helphaverErr := tx.Exec(ctx, `DELETE FROM membership_helphaver WHERE membership_id=$1`, membershipID)

	if membership_manualErr != nil || membership_specialErr != nil || membership_automaticErr != nil || membership_helphaverErr != nil {
		return fmt.Errorf("problem deleting membership sub table: %w", membership_manualErr)
	}

	// commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("problem committing transaction: %w", commitErr)
	}

	return nil
}

func (db *ProfileDB) GetMultipleMembership(ctx context.Context, intSkip int, intLimit int, month int, year int, userID string) ([]Membership, error) {
	memberships := []Membership{}

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
		return []Membership{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r Membership
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
			return []Membership{}, err
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

func prepareMembershipUpdateQuery(req Membership) (string, []interface{}) {
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
