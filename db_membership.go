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

type order struct {
	ID          string `json:"ID"`
	Status      string `json:"Status"`
	ProductType string `json:"ProductType"`
}

type orderRes struct {
	Message string  `json:"message"`
	Success bool    `json:"success"`
	Data    []order `json:"data"`
}

type orderDeleteRes struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    int    `json:"data"`
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

func (db *pgProfileDB) cancelMembership(ctx context.Context, membBody membershipCancellationBody) (error, int, int, int) {

	var email string
	var user_id string

	if membBody.UserID != nil {
		if err := db.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, *membBody.UserID).Scan(&email); err != nil {
			return fmt.Errorf("error while getting user email: %w", err), 0, 0, 0
		}
		user_id = *membBody.UserID
	} else if membBody.KeycloakID != nil {
		if err := db.QueryRow(ctx, `SELECT email, user_id FROM users WHERE keycloak_id=$1`, *membBody.KeycloakID).Scan(&email, &user_id); err != nil {
			return fmt.Errorf("error while getting user email: %w", err), 0, 0, 0
		}
	} else {
		if err := db.QueryRow(ctx, `SELECT user_id FROM users WHERE email=$1`, *membBody.Email).Scan(&user_id); err != nil {
			return fmt.Errorf("error while getting user id: %w", err), 0, 0, 0
		}
		email = *membBody.Email
	}

	authHeader := ctx.Value("Authorization").(string)

	userOrderDetails := "https://api.eurokab.info/pay/v2/orders?email=" + email + "&product-type=globalmembership"

	orderDetails := utils.HTTPCallAndGetBody(userOrderDetails, authHeader, nil, "GET")

	// un marshal order details
	var orderDetailsRes orderRes
	err := json.Unmarshal([]byte(orderDetails), &orderDetailsRes)
	if err != nil {
		return fmt.Errorf("error while unmarshalling order details: %w", err), 0, 0, 0
	}

	// loop over order details and cancel the order
	for _, order := range orderDetailsRes.Data {
		cancelOrder := "https://api.eurokab.info/pay/v2/order/" + order.ID
		postBody, _ := json.Marshal(map[string]interface{}{
			"Status": "cancelled",
		})

		buffPostBody := bytes.NewBuffer(postBody)

		cancelOrderRes := utils.HTTPCallAndGetBody(cancelOrder, authHeader, buffPostBody, "PATCH")

		// un marshal cancel order details
		var cancelOrderResRes orderRes
		err := json.Unmarshal([]byte(cancelOrderRes), &cancelOrderResRes)
		if err != nil {
			return fmt.Errorf("error while unmarshalling cancel order details: %w", err), 0, 0, 0
		}
	}

	// update user grant cancelled_at to now
	grantUpdateRes, grantUpdateErr := db.Exec(ctx, `UPDATE "grant" SET cancelled_at=$1 WHERE user_id=$2`, time.Now(), user_id)

	if grantUpdateErr != nil {
		return fmt.Errorf("error while updating grant: %w", grantUpdateErr), 0, 0, 0
	}

	numberOfRowsUpdated := grantUpdateRes.RowsAffected()

	specialTableDelete := "https://api.eurokab.info/pay/v2/special/" + email

	specialTableDelRes := utils.HTTPCallAndGetBody(specialTableDelete, authHeader, nil, "DELETE")

	// un marshal special table delete details
	var specialTableDelResRes orderDeleteRes

	err = json.Unmarshal([]byte(specialTableDelRes), &specialTableDelResRes)

	if err != nil {
		return fmt.Errorf("error while unmarshalling special table delete details: %w", err), 0, 0, 0
	}

	return nil, len(orderDetailsRes.Data), int(numberOfRowsUpdated), specialTableDelResRes.Data

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
