package repo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

type membershipInterface interface {
	GetMembershipByID(ctx context.Context, id int) (Membership, error)
	GetMembershipByUserID(ctx context.Context, userID string) (UserMembershipRes, error)
	GetMembershipByKCID(ctx context.Context, kcID string) (UserMembershipRes, error)
	GetMultipleMembership(ctx context.Context, intSkip int, intLimit int, userID string) ([]Membership, error)
	GetExpiredMemberships(ctx context.Context, intSkip int, intLimit int) ([]Membership, error)
	PatchMembershipByID(ctx context.Context, membership Membership, id int) (int, error)
	SoftDeleteMembershipByID(ctx context.Context, id int) error
	CancelMembership(ctx context.Context, body EmailKeycloakAndUserIDBody) error
	GetAutomaticMembershipByMembershipID(ctx context.Context, membershipID int) (MembershipAutomatic, error)
	EvaluateMembershipByUserID(ctx context.Context, evalbody EmailKeycloakAndUserIDBody) (UserMembershipRes, error)
}

type Membership struct {
	ID        *int       `json:"id" db:"id"`
	Active    *bool      `json:"active"`
	UserID    *uuid.UUID `json:"user_id"`
	Type      *string    `json:"type"`
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
			PaymentType   *string    `json:"payment_type,omitempty"`
			Status        *string    `json:"status,omitempty"`
		} `json:"payment,omitempty"`
		Automatic struct {
			OrderID   *int `json:"order_id,omitempty"`
			PaymentID *int `json:"payment_id,omitempty"`
		} `json:"automatic,omitempty"`
		Manual struct {
			OrderID   *int `json:"order_id,omitempty"`
			PaymentID *int `json:"payment_id,omitempty"`
			Quantity  *int `json:"quantity,omitempty"`
		} `json:"manual,omitempty"`
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
	Slug      *string                 `json:"slug"`
	Content   *map[string]interface{} `json:"content,omitempty"`
	CreatedAt *time.Time              `json:"created_at"`
}

type EmailKeycloakAndUserIDBody struct {
	Email      *string `json:"email"`
	KeycloakID *string `json:"keycloak_id"`
	UserID     *string `json:"user_id"`
}

func (db *ProfileDB) EvaluateMembershipByUserID(ctx context.Context, evalBody EmailKeycloakAndUserIDBody) (UserMembershipRes, error) {
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

	var membershipInsertData = Membership{
		Active: utils.PointerBool(false),
	}

	var currentMembership string
	var latestOrderPaymentDate *time.Time
	var latestOrderPaymentID *int
	var latestOrderPaymentStatus string
	var latestOrder orders.Order
	var allOrderCancelled bool
	var userInSpecialTable bool
	var orderStartingDate time.Time
	var orderQuantity int
	var lastApprovedRequest *RequestAndGrant

	const LIMIT = 200

	var uuidErr error
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

	// fetch all the order of the user ( limit 200 as of now )
	ordersService := db.ordersServiceFactory()
	userOrders, err := ordersService.GetOrders(ctx, email, "globalmembership", true, "desc", LIMIT, 0)
	if err != nil {
		return UserMembershipRes{}, fmt.Errorf("ordersService.GetOrders: %w", err)
	}

	approvedRequests, err := db.GetMultipleRequest(ctx, 0, 1, userKeycloakID, common.RequestStatusApproved, "", common.RequestTypeHelpHaver, "desc")
	if err != nil {
		return UserMembershipRes{}, fmt.Errorf("db.GetMultipleRequest: %w", err)
	}
	if len(approvedRequests) > 0 {
		lastApprovedRequest = &approvedRequests[0]
	}

	// Evaluation starts here

	if len(userOrders) != 0 {

		// count number of cancelled order in orderDetailsRes.Data
		var cancelledOrderCount int
		for _, order := range userOrders {
			if order.Status == "cancelled" {
				cancelledOrderCount++
			}
		}

		// if all orders are cancelled, then current membership is cancelled
		if cancelledOrderCount != 0 && cancelledOrderCount == len(userOrders) {
			allOrderCancelled = true
		} else {
			// latest order
			latestOrder = userOrders[0]
			latestOrderPaymentDate = &latestOrder.PaymentDate

			if latestOrder.Type == "recurring" {
				currentMembership = "automatic"
			} else {
				currentMembership = "manual"
			}

			// fetch latest order payment
			orderPayments, err := ordersService.GetOrderPayments(ctx, latestOrder.ID, "desc", 1, 0)
			if err != nil {
				return UserMembershipRes{}, fmt.Errorf("ordersService.GetOrderPayments: %w", err)
			}

			if len(orderPayments) == 0 {
				fmt.Printf("payment details not found for order id: %v", latestOrder.ID)
				// set latestOrderPaymentID as -1
				latestOrderPaymentID = utils.PointerInt(-1)
				// TODO: analyse and set the latestOrderPaymentStatus as well
			} else {
				latestOrderPaymentID = &orderPayments[0].ID
				latestOrderPaymentStatus = orderPayments[0].Status
			}

			// Regular (manual) orders are relative with regard to the timeline.
			// We have to align them to each other and "attach" them to an absolute date.
			// The combination of "StartingDate" and "Quantity"
			// We do this for paid only

			// filter userOrders to exclude cancelled orders
			var paidManualOrders []orders.Order
			for _, order := range userOrders {
				if order.Type == "regular" && order.Status == "paid" {
					paidManualOrders = append(paidManualOrders, order)
				}
			}

			var previousStartingDate time.Time
			var previousOrderQuantity int
			for i := len(paidManualOrders) - 1; i >= 0; i-- {
				order := paidManualOrders[i]
				orderQuantity = utils.Max(1, order.Quantity)

				if !order.StartingDate.IsZero() {
					orderStartingDate = order.StartingDate
					previousStartingDate = orderStartingDate
					previousOrderQuantity = orderQuantity
					continue
				}

				// first order
				if i == len(paidManualOrders)-1 {
					orderStartingDate = order.PaymentDate
				} else {
					if order.PaymentDate.Before(previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)) {
						orderStartingDate = previousStartingDate.AddDate(0, 0, 30*previousOrderQuantity)
					} else {
						orderStartingDate = order.PaymentDate
					}
				}

				previousStartingDate = orderStartingDate
				previousOrderQuantity = orderQuantity

				// update order starting date
				if err := ordersService.SetOrderStartingDate(ctx, order.ID, orderStartingDate); err != nil {
					return UserMembershipRes{},
						fmt.Errorf("error updating order [%d] starting date: %w", order.ID, err)
				}
			}
		}
	}

	if lastApprovedRequest != nil {
		if latestOrderPaymentDate != nil {
			if latestOrderPaymentDate.After(*lastApprovedRequest.Grant.CreatedAt) {
				// cancel grant
				lastApprovedRequest.Grant.CancelledAt = utils.PointerTime(time.Now())
				_, err := db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE id=$2`,
					*lastApprovedRequest.Grant.CancelledAt, *lastApprovedRequest.Grant.ID)
				if err != nil {
					return UserMembershipRes{}, fmt.Errorf("problem updating grant [%d]: %w", *lastApprovedRequest.Grant.ID, err)
				}
			} else {
				currentMembership = "helphaver"
			}
		} else {
			currentMembership = "helphaver"
		}
	}

	if currentMembership == "automatic" {
		if time.Now().AddDate(0, 0, -60).Before(*latestOrderPaymentDate) {
			membershipInsertData.Active = utils.PointerBool(true)
		}
	} else if currentMembership == "manual" {
		membershipInsertData.Expiry = utils.PointerTime(orderStartingDate.AddDate(0, 0, 30*orderQuantity))
		if time.Now().AddDate(0, 0, -60).Before(*membershipInsertData.Expiry) {
			membershipInsertData.Active = utils.PointerBool(true)
		}
	} else if currentMembership == "helphaver" {
		var props map[string]interface{}
		if lastApprovedRequest.Grant.Properties.Valid {
			if err := lastApprovedRequest.Grant.Properties.Unmarshal(&props); err != nil {
				return UserMembershipRes{}, fmt.Errorf("json.Unmarshal grant properties [%d]: %w", *lastApprovedRequest.Grant.ID, err)
			}
		} else {
			return UserMembershipRes{}, fmt.Errorf("empty grant properties [%d]", *lastApprovedRequest.Grant.ID)
		}

		if months, ok := props["months"]; ok {
			if monthsVal, ok := months.(float64); ok {
				membershipInsertData.Expiry = utils.PointerTime(lastApprovedRequest.Grant.CreatedAt.AddDate(0, 0, 30*int(monthsVal)))
			} else {
				return UserMembershipRes{}, fmt.Errorf("malformed grant property 'months' [%d]", *lastApprovedRequest.Grant.ID)
			}
		} else {
			return UserMembershipRes{}, fmt.Errorf("missing grant property 'months' [%d]", *lastApprovedRequest.Grant.ID)
		}

		if time.Now().AddDate(0, 0, -60).Before(*membershipInsertData.Expiry) {
			membershipInsertData.Active = utils.PointerBool(true)
		}
	}

	if membershipInsertData.Expiry == nil || membershipInsertData.Expiry.Before(time.Now()) {
		special, err := ordersService.GetSpecial(ctx, email)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("ordersService.GetSpecial: %w", err)
		}

		if special != nil {
			userInSpecialTable = true
			currentMembership = "special"
			membershipInsertData.Active = utils.PointerBool(true)
			membershipInsertData.Type = utils.PointerString("special")
			membershipInsertData.Expiry = nil //TODO (edo): this should probably be changed to have dates
		}
	}

	// check currentMembership is cancelled or new
	if !userInSpecialTable {
		// We expect all orders before a grant to be cancelled.
		// Note that a grant is cancelled above if an order was payed after the grant

		if len(userOrders) == 0 && lastApprovedRequest == nil {
			// never had an order and never had a grant
			currentMembership = "new"
		} else if allOrderCancelled && (lastApprovedRequest == nil || lastApprovedRequest.Grant.CancelledAt != nil) {
			// all orders are cancelled and either never had a grant or last grant was cancelled
			currentMembership = "cancelled"
		} else if lastApprovedRequest != nil && lastApprovedRequest.Grant.CancelledAt != nil && len(userOrders) == 0 {
			// last grant was cancelled and never had an order
			currentMembership = "cancelled"
		}
	}
	membershipInsertData.Type = &currentMembership

	// check if user has an existing membership
	userMemb, userMembErr := db.GetMultipleMembership(ctx, 0, 100, userID)
	if userMembErr != nil {
		return UserMembershipRes{}, fmt.Errorf("error getting user membership: %w", userMembErr)
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
			return UserMembershipRes{}, fmt.Errorf("error inserting membership: %w", insertMembErr)
		}
	} else {
		// update membership
		membershipID, updateMembErr = db.PatchMembershipByID(ctx, membershipInsertData, *userMemb[0].ID)
		if updateMembErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error updating membership: %w", updateMembErr)
		}

		// remove all the previous sub memberships entries as they'll be added again ( latest will be added )
		deleteAllMembershipSubTableErr := db.deleteAllMembershipSubTableByMembershipID(ctx, membershipID)
		if deleteAllMembershipSubTableErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error deleting membership sub tables: %w", deleteAllMembershipSubTableErr)
		}
	}

	if currentMembership == "automatic" {
		_, automaticErr := db.createAutomaticMembership(ctx, MembershipAutomatic{
			MembershipID: &membershipID,
			OrderID:      &latestOrder.ID,
			PaymentID:    latestOrderPaymentID,
		})

		if automaticErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error creating automatic membership [%d]: %w", membershipID, automaticErr)
		}
	} else if currentMembership == "manual" {
		_, manualErr := db.createManualMembership(ctx, MembershipManual{
			MembershipID: &membershipID,
			OrderID:      &latestOrder.ID,
			PaymentID:    latestOrderPaymentID,
			Quantity:     &latestOrder.Quantity,
		})

		if manualErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error creating manual membership [%d]: %w", membershipID, manualErr)
		}
	} else if currentMembership == "helphaver" {
		_, helphaverErr := db.createHelpHaverMembership(ctx, MembershipHelpHaver{
			MembershipID: &membershipID,
			GrantID:      lastApprovedRequest.Grant.ID,
		})

		if helphaverErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error creating helphaver membership [%d]: %w", membershipID, helphaverErr)
		}
	} else if currentMembership == "special" {
		_, specialErr := db.createSpecialMembership(ctx, MembershipSpecial{
			MembershipID: &membershipID,
		})

		if specialErr != nil {
			return UserMembershipRes{}, fmt.Errorf("error creating special membership [%d]: %w", membershipID, specialErr)
		}
	}

	// add user notification
	var notificationSlugs []string

	if currentMembership == "automatic" && latestOrderPaymentStatus == "nosuccess" {
		notificationSlugs = append(notificationSlugs, common.NotificationSlugMBProblemPreviousPayment)
		if time.Now().AddDate(0, 0, -60).After(*latestOrderPaymentDate) {
			notificationSlugs = append(notificationSlugs, common.NotificationSlugMBHasExpiredNotice)
		}
	} else if (currentMembership == "manual" || currentMembership == "helphaver") &&
		time.Now().After(*membershipInsertData.Expiry) {
		if time.Now().AddDate(0, 0, -60).After(*membershipInsertData.Expiry) {
			notificationSlugs = append(notificationSlugs, common.NotificationSlugMBHasExpiredNotice)
		} else {
			notificationSlugs = append(notificationSlugs, common.NotificationSlugMBExpirationNotice)
		}
	} else if currentMembership == "cancelled" {
		notificationSlugs = append(notificationSlugs, common.NotificationSlugMBCancelled)
	} else if currentMembership == "new" {
		notificationSlugs = append(notificationSlugs, common.NotificationSlugMBNew)
	}

	// Deactivate previous user notifications
	updateAllUserNotificationToInactiveErr := db.deactivateUserNotifications(ctx, userID, allMembershipNotifications...)
	if updateAllUserNotificationToInactiveErr != nil {
		return UserMembershipRes{}, fmt.Errorf("error updating all user notification to inactive: %w", updateAllUserNotificationToInactiveErr)
	}

	if len(notificationSlugs) != 0 {
		// loop through all the slugs and add notification
		for _, slug := range notificationSlugs {
			err := db.CreateUserNotification(ctx, UserNotification{
				UserID:         &userID,
				NotificationID: NotificationsRegistry.BySlug[slug].ID,
				Active:         utils.PointerBool(true),
				SeenAt:         nil,
			})

			if err != nil {
				return UserMembershipRes{}, fmt.Errorf("db.CreateUserNotification [%s]: %w", slug, err)
			}
		}
	}

	userMembershipResponse, userMembershipResponseErr := db.GetMembershipByUserID(ctx, userID)
	if userMembershipResponseErr != nil {
		return UserMembershipRes{}, fmt.Errorf("error getting membership [%s]: %w", userID, userMembershipResponseErr)
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

func (db *ProfileDB) GetMembershipByKCID(ctx context.Context, kcID string) (UserMembershipRes, error) {
	var userID uuid.UUID
	if err := db.QueryRow(ctx, "SELECT user_id FROM users WHERE keycloak_id=$1", kcID).
		Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserMembershipRes{}, common.ErrNotFound
		}
		return UserMembershipRes{}, fmt.Errorf("error while getting membership: %w", err)
	}

	return db.GetMembershipByUserID(ctx, userID.String())
}

func (db *ProfileDB) GetMembershipByUserID(ctx context.Context, userID string) (UserMembershipRes, error) {
	var membership UserMembershipRes

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		active,
		user_id,
		type,
		expiry,
		created_at,
		updated_at,
		deleted_at
		FROM membership 
		WHERE user_id = $1`, userID).Scan(
		&membership.ID,
		&membership.Active,
		&membership.UserID,
		&membership.Type,
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

	ordersService := db.ordersServiceFactory()

	if *membership.Type == "automatic" {
		autoMembership, err := db.GetAutomaticMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting automatic membership: %w", err)
		}
		membership.Details.Automatic.OrderID = autoMembership.OrderID
		membership.Details.Automatic.PaymentID = autoMembership.PaymentID

		payment, err := ordersService.GetPaymentByID(ctx, *autoMembership.PaymentID)
		if err != nil {
			return UserMembershipRes{},
				fmt.Errorf("ordersService.GetPaymentByID [%d]: %w", *autoMembership.PaymentID, err)
		}

		membership.Details.Payment.Amount = &payment.Amount
		membership.Details.Payment.Currency = &payment.Currency
		membership.Details.Payment.Status = &payment.PaymentStatus
		membership.Details.Payment.Date = &payment.CreatedAt
		membership.Details.Payment.PaymentMethod = &payment.CCNumber
		membership.Details.Payment.PaymentType = &payment.PaymentType

	} else if *membership.Type == "helphaver" {
		helphaverMembership, err := db.getHelphaverMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting helphaver membership: %w", err)
		}
		membership.Details.HelpHaver.CreatedAt = helphaverMembership.CreatedAt
		membership.Details.HelpHaver.NbMonths = helphaverMembership.NbMonths
	} else if *membership.Type == "manual" {
		manualMembership, err := db.getManualMembershipByMembershipID(ctx, *membership.ID)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting manual membership: %w", err)
		}

		membership.Details.Manual.OrderID = manualMembership.OrderID
		membership.Details.Manual.PaymentID = manualMembership.PaymentID
		membership.Details.Manual.Quantity = manualMembership.Quantity

		payment, err := ordersService.GetPaymentByID(ctx, *manualMembership.PaymentID)
		if err != nil {
			return UserMembershipRes{},
				fmt.Errorf("ordersService.GetPaymentByID [%d]: %w", *manualMembership.PaymentID, err)
		}

		membership.Details.Payment.Amount = &payment.Amount
		membership.Details.Payment.Currency = &payment.Currency
		membership.Details.Payment.Status = &payment.PaymentStatus
		membership.Details.Payment.Date = &payment.CreatedAt
		membership.Details.Payment.PaymentMethod = &payment.CCNumber
		membership.Details.Payment.PaymentType = &payment.PaymentType

	} else if *membership.Type == "special" {
		var email string
		if err := db.QueryRow(ctx, `SELECT primary_email FROM users WHERE user_id=$1`, userID).Scan(&email); err != nil {
			return UserMembershipRes{}, fmt.Errorf("error while getting User email: %w", err)
		}

		special, err := ordersService.GetSpecial(ctx, email)
		if err != nil {
			return UserMembershipRes{}, fmt.Errorf("ordersService.GetSpecial: %w", err)
		}

		if special != nil {
			membership.Details.Special.ApprovedBy = &special.Category
			membership.Details.Special.Type = &special.SubCategory
		} else {
			log.Printf("WARNING: user no longer in special table %s\n", email)
		}
	}

	// fetch active user notification
	userActiveNotification, userNotiErr := db.GetActiveUserNotificationByUserID(ctx, userID)

	if userNotiErr != nil {
		return UserMembershipRes{}, fmt.Errorf("error while getting active user notification: %w", userNotiErr)
	}

	var userNotiSlug []UserMembershipNotification
	// loop and add slug to membership
	for _, noti := range userActiveNotification {
		var userNoti UserMembershipNotification
		userNoti.Slug = noti.Slug
		userNoti.Content = noti.Content
		userNoti.CreatedAt = noti.CreatedAt
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
		mh.id,
		mh.grant_id,
		mh.membership_id,
		(g.properties->'months')::integer,
		mh.created_at,
		mh.updated_at,
		mh.deleted_at 
		from membership_helphaver mh LEFT JOIN "grant" g ON mh.grant_id = g.id 
		WHERE mh.membership_id = $1`, membershipID).Scan(
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

func (db *ProfileDB) CancelMembership(ctx context.Context, membBody EmailKeycloakAndUserIDBody) error {
	var email string
	var user_id string

	if membBody.UserID != nil && *membBody.UserID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email FROM users WHERE user_id=$1`, *membBody.UserID).Scan(&email); err != nil {
			return fmt.Errorf("error while getting user email: %w", err)
		}
		user_id = *membBody.UserID
	} else if membBody.KeycloakID != nil && *membBody.KeycloakID != "" {
		if err := db.QueryRow(ctx, `SELECT primary_email, user_id FROM users WHERE keycloak_id=$1`, *membBody.KeycloakID).Scan(&email, &user_id); err != nil {
			return fmt.Errorf("error while getting user email: %w", err)
		}
	} else {
		if err := db.QueryRow(ctx, `SELECT user_id FROM users WHERE primary_email=$1`, *membBody.Email).Scan(&user_id); err != nil {
			return fmt.Errorf("error while getting user id: %w", err)
		}
		email = *membBody.Email
	}

	tx, txErr := db.Begin(ctx)
	if txErr != nil {
		return txErr
	}

	defer func() { _ = tx.Rollback(ctx) }()

	ordersService := db.ordersServiceFactory()

	// cancel orders
	userOrders, err := ordersService.GetOrders(ctx, email, "globalmembership", false, "", 100, 0)
	if err != nil {
		return fmt.Errorf("ordersService.GetOrders: %w", err)
	}

	for _, order := range userOrders {
		if order.Status == "cancelled" {
			continue
		}
		if err := ordersService.CancelOrder(ctx, order.ID); err != nil {
			return fmt.Errorf("ordersService.CancelOrder [%d]: %w", order.ID, err)
		}
	}

	// delete special
	if err = ordersService.DeleteSpecial(ctx, email); err != nil {
		return fmt.Errorf("ordersService.DeleteSpecial : %w", err)
	}

	// cancel grants
	_, err = db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE user_id=$2 AND type LIKE 'mb_%' AND cancelled_at IS NULL`, time.Now(), user_id)
	if err != nil {
		return fmt.Errorf("error while updating grant: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	_, err = db.EvaluateMembershipByUserID(ctx, membBody)
	if err != nil {
		return fmt.Errorf("db.EvaluateMembershipByUserID: %w", err)
	}

	return nil
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

func (db *ProfileDB) GetMultipleMembership(ctx context.Context, intSkip int, intLimit int, userID string) ([]Membership, error) {
	memberships := []Membership{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereMembershipQuery(userID)

	rows, err := db.Query(ctx, `
		SELECT 
		id,
		active,
		user_id,
		type,
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
			&r.Expiry,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []Membership{}, err
		}

		memberships = append(memberships, r)
	}
	if err = rows.Err(); err != nil {
		return []Membership{}, fmt.Errorf("rows.Err: %w", err)
	}

	return memberships, nil
}

func (db *ProfileDB) GetExpiredMemberships(ctx context.Context, intSkip int, intLimit int) ([]Membership, error) {
	rows, err := db.Query(ctx, `
		SELECT 
			id, active, user_id, type, expiry, created_at, updated_at, deleted_at
		FROM membership 
		WHERE active = true AND expiry IS NOT NULL AND expiry < $1
		LIMIT $2 OFFSET $3`, time.Now().UTC(), intLimit, intSkip)
	if err != nil {
		return []Membership{}, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	memberships := make([]Membership, 0)
	for rows.Next() {
		var r Membership
		if err := rows.Scan(
			&r.ID,
			&r.Active,
			&r.UserID,
			&r.Type,
			&r.Expiry,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []Membership{}, fmt.Errorf("rows.Scan: %w", err)
		}

		memberships = append(memberships, r)
	}
	if err = rows.Err(); err != nil {
		return []Membership{}, fmt.Errorf("rows.Err: %w", err)
	}

	return memberships, nil
}

func buildAndGetWhereMembershipQuery(userID string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

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
