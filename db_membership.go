package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	ID          int    `json:"ID"`
	Status      string `json:"Status"`
	ProductType string `json:"ProductType"`
}

type payment struct {
	Amount        int       `json:"Amount"`
	DebitCurrency string    `json:"DebitCurrency"`
	CCNumber      string    `json:"CCNumber"`
	PaymentStatus string    `json:"PaymentStatus"`
	CreatedAt     time.Time `json:"created_at"`
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

type orderRes struct {
	MessageAndSuccess
	Data []order `json:"data"`
}

type orderDeleteRes struct {
	MessageAndSuccess
	Data int `json:"data"`
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

		orderDetails := utils.HTTPCallAndGetBody(userPaymentDetails, authHeader, nil, "GET")

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

		paymentDetails := utils.HTTPCallAndGetBody(manualUserPaymentDetails, authHeader, nil, "GET")

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

		speMemDetails := utils.HTTPCallAndGetBody(specialMembDetailUrl, authHeader, nil, "GET")

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

func (db *pgProfileDB) cancelMembership(ctx context.Context, membBody membershipCancellationBody, authHeader string) (int, int, int, error) {

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

	userOrderDetails := getServerUrl() + "/pay/v2/orders?email=" + email + "&product-type=globalmembership"

	orderDetails := utils.HTTPCallAndGetBody(userOrderDetails, authHeader, nil, "GET")

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

		cancelOrderRes := utils.HTTPCallAndGetBody(cancelOrder, authHeader, buffPostBody, "PATCH")

		// un marshal cancel order details
		var cancelOrderResRes orderRes
		err := json.Unmarshal(cancelOrderRes, &cancelOrderResRes)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("error while unmarshalling cancel order details: %w", err)
		}
	}

	// update user grant cancelled_at to now
	grantUpdateRes, grantUpdateErr := db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE user_id=$2`, time.Now(), user_id)

	if grantUpdateErr != nil {
		return 0, 0, 0, fmt.Errorf("error while updating grant: %w", grantUpdateErr)
	}

	numberOfRowsUpdated := grantUpdateRes.RowsAffected()

	specialTableDelete := getServerUrl() + "/pay/v2/special/" + email

	specialTableDelRes := utils.HTTPCallAndGetBody(specialTableDelete, authHeader, nil, "DELETE")

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

func (db *pgProfileDB) getMultipleMembership(ctx context.Context, intSkip int, intLimit int) ([]membershipRes, error) {
	memberships := []membershipRes{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereMembershipQuery()

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

func buildAndGetWhereMembershipQuery() (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// Add where conditions when required

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
