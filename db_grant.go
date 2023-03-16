package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

func (db *pgProfileDB) getGrantByIDAndUserID(ctx context.Context, id int, userID string) (grantRes, error) {
	var grant grantRes
	var whereQuery string

	if id != 0 {
		whereQuery = fmt.Sprintf(" WHERE id=%d", id)
	} else if userID != "" {
		whereQuery = fmt.Sprintf(" WHERE user_id='%s'", userID)
	} else {
		return grantRes{}, fmt.Errorf("invalid values")
	}

	if err := db.QueryRow(ctx, `
	SELECT id,
		user_id,
		amount,
		currency,
		type,
		loaned,
		granted,
		repayed,
		request_id,
		cancelled_at,
		created_at,
		updated_at,
		deleted_at 
	FROM "grant"`+whereQuery).Scan(
		&grant.ID,
		&grant.UserID,
		&grant.Amount,
		&grant.Currency,
		&grant.Type,
		&grant.Loaned,
		&grant.Granted,
		&grant.Repayed,
		&grant.RequestID,
		&grant.CancelledAt,
		&grant.CreatedAt,
		&grant.UpdatedAt,
		&grant.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return grantRes{}, errNotFound
		}
		return grantRes{}, fmt.Errorf("error while getting grant: %w", err)
	}

	return grant, nil
}

func (db *pgProfileDB) createGrant(ctx context.Context, req grantAndGrantMembership) (int, error) {

	if *req.Type == "hhmembership" {
		if req.Amount == nil {
			// set default req.Amount to 10
			var defaultAmount = new(int)
			*defaultAmount = 10
			totalAmount := *defaultAmount * *req.Month
			req.Amount = &totalAmount
		}

		if req.Currency == nil {
			var defaultCurrency = new(string)
			*defaultCurrency = "EUR"
			req.Currency = defaultCurrency
		}
	}

	var ID int

	createString, numString, createQueryArgs := prepareGrantCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO "grant" (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) createGrantMembership(ctx context.Context, req grantAndGrantMembership) (int, error) {

	var ID int

	createString, numString, createQueryArgs := prepareGrantMembershipCreateQuery(req)

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO grant_membership (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem creating grant: %w", err)
		}

		return ID, nil

	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) patchGrantMembership(ctx context.Context, req grantMembeship, grantId int) (int, error) {

	var ID int

	toUpdate, toUpdateArgs := prepareGrantMembershipUpdate(req)

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE grant_membership SET %s WHERE grant_id='%d' RETURNING id`, toUpdate, grantId),
			toUpdateArgs...).
			Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem updating grant_membership: %w", err)
		}

		return ID, nil
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) getGrantMembershipByGrantID(ctx context.Context, grantId int) (grantMembeshipRes, error) {

	var grantMemb grantMembeshipRes

	if err := db.QueryRow(ctx, `
		SELECT 
		id,
		grant_id,
		nb_months,
		months_used,
		months_left,
		created_at,
		updated_at,
		deleted_at
		FROM grant_membership WHERE grant_id=$1`, grantId).Scan(
		&grantMemb.ID,
		&grantMemb.GrantID,
		&grantMemb.Month,
		&grantMemb.MonthsUsed,
		&grantMemb.MonthsLeft,
		&grantMemb.CreatedAt,
		&grantMemb.UpdatedAt,
		&grantMemb.DeletedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return grantMembeshipRes{}, errNotFound
		}
		return grantMembeshipRes{}, fmt.Errorf("error while getting grant membership: %w", err)
	}

	return grantMemb, nil
}

func (db *pgProfileDB) patchGrant(ctx context.Context, grant grant, id int, requestID int) (int, error) {

	var whereQuery string

	if requestID != 0 {
		whereQuery = fmt.Sprintf("WHERE request_id=%d", requestID)
	} else if id != 0 {
		whereQuery = fmt.Sprintf("WHERE id=%d", id)
	} else {
		return 0, fmt.Errorf("invalid values")
	}

	if *grant.Type == "hhmembership" {

		if grant.Amount == nil {
			// set default req.Amount to 10
			var defaultAmount = new(int)
			*defaultAmount = 10
			grant.Amount = defaultAmount
		}

		if grant.Currency == nil {
			var defaultCurrency = new(string)
			*defaultCurrency = "EUR"
			grant.Currency = defaultCurrency
		}
	}

	toUpdate, toUpdateArgs := prepareGrantUpdate(grant)

	var grantID int

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE "grant" SET %s %s RETURNING id`, toUpdate, whereQuery),
			toUpdateArgs...).Scan(&grantID); err != nil {
			if err == pgx.ErrNoRows {
				return 0, errNotFound
			}
			return 0, fmt.Errorf("problem updating grant: %w", err)
		}

		return grantID, nil
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) softDeleteGrantByID(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `UPDATE "grant" SET deleted_at=$1 WHERE id=$2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("problem soft deleting grant: %w", err)
	}
	return nil
}

func (db *pgProfileDB) getNumberOfGrantMonthsUsedByGrantID(ctx context.Context, grantID int) (int, error) {

	var monthUsed *int

	if err := db.QueryRow(ctx, `
	SELECT COUNT(*)
    FROM membership_helphaver
    WHERE grant_id = $1
	`, grantID).Scan(
		&monthUsed,
	); err != nil {
		return 0, fmt.Errorf("error while getting number of months used in grant by grant ID: %w", err)
	}

	return *monthUsed, nil
}

func (db *pgProfileDB) getMultipleGrant(ctx context.Context, intSkip int, intLimit int, cancelled *bool, userID string, grantType string, createdAt string) ([]grantRes, error) {
	grants := []grantRes{}

	userDbWhereQuery, orderByQuery := buildAndGetWhereGrantQuery(cancelled, userID, grantType, createdAt)

	rows, err := db.Query(ctx, `
		SELECT 
		id,
		user_id,
		amount,
		currency,
		type,
		loaned,
		granted,
		repayed,
		request_id,
		cancelled_at,
		created_at,
		updated_at,
		deleted_at 
		FROM "grant" `+userDbWhereQuery+
		orderByQuery+
		" LIMIT $1 OFFSET $2", intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []grantRes{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r grantRes
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.Amount,
			&r.Currency,
			&r.Type,
			&r.Loaned,
			&r.Granted,
			&r.Repayed,
			&r.RequestID,
			&r.CancelledAt,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []grantRes{}, err
		}

		grants = append(grants, r)
	}

	return grants, nil
}

func buildAndGetWhereGrantQuery(cancelled *bool, userID string, grantType string, createdAt string) (string, string) {

	var whereString strings.Builder
	var orderBy strings.Builder
	var whereCondition strings.Builder
	whereString.WriteString(" WHERE")
	whereCondition.WriteString("")

	// Add where conditions when required

	if cancelled != nil {
		if whereCondition.String() != "" {
			if *cancelled {
				whereCondition.WriteString(" AND cancelled_at IS NOT NULL")
			} else {
				whereCondition.WriteString(" AND cancelled_at IS NULL")
			}
		} else {
			if *cancelled {
				whereCondition.WriteString(" cancelled_at IS NOT NULL")
			} else {
				whereCondition.WriteString(" cancelled_at IS NULL")
			}
		}
	}

	if grantType != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND type='%s'", grantType))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" type='%s'", grantType))
		}
	}

	if userID != "" {
		if whereCondition.String() != "" {
			whereCondition.WriteString(fmt.Sprintf(" AND user_id='%s'", userID))
		} else {
			whereCondition.WriteString(fmt.Sprintf(" user_id='%s'", userID))
		}
	}

	if createdAt != "" {
		if strings.ToLower(createdAt) != "desc" && strings.ToLower(createdAt) != "asc" {
			createdAt = "asc"
		}
		orderBy.WriteString(fmt.Sprintf(" ORDER BY created_at %s", createdAt))
	} else {
		orderBy.WriteString(fmt.Sprintf(" ORDER BY updated_at %s", "desc"))
	}

	if whereCondition.String() != "" {
		whereString.WriteString(whereCondition.String())
	} else {
		whereString.Reset()
	}
	return whereString.String(), orderBy.String()
}

func prepareGrantCreateQuery(req grantAndGrantMembership) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Amount != nil {
		createStrings = append(createStrings, "amount")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Amount)
	}
	if req.Currency != nil {
		createStrings = append(createStrings, "currency")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Currency)
	}
	if req.UserID != nil {
		createStrings = append(createStrings, "user_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.UserID)
	}
	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}
	if req.Loaned != nil {
		createStrings = append(createStrings, "loaned")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Loaned)
	}
	if req.Granted != nil {
		createStrings = append(createStrings, "granted")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Granted)
	}
	if req.Repayed != nil {
		createStrings = append(createStrings, "repayed")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Repayed)
	}
	if req.RequestID != nil {
		createStrings = append(createStrings, "request_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestID)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareGrantMembershipCreateQuery(req grantAndGrantMembership) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.GrantID != nil {
		createStrings = append(createStrings, "grant_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.GrantID)
	}
	if req.Month != nil {
		createStrings = append(createStrings, "nb_months")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Month)
	}
	if req.MonthsLeft != nil {
		createStrings = append(createStrings, "months_left")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MonthsLeft)
	}
	if req.MonthsUsed != nil {
		createStrings = append(createStrings, "months_used")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.MonthsUsed)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareGrantMembershipUpdate(req grantMembeship) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Month != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("nb_months=$%d", len(updateStrings)+1))
		args = append(args, *req.Month)
	}
	if req.MonthsLeft != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("months_left=$%d", len(updateStrings)+1))
		args = append(args, *req.MonthsLeft)
	}
	if req.MonthsUsed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("months_used=$%d", len(updateStrings)+1))
		args = append(args, *req.MonthsUsed)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}

func prepareGrantUpdate(req grant) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Amount != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("amount=$%d", len(updateStrings)+1))
		args = append(args, *req.Amount)
	}
	if req.UserID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("user_id=$%d", len(updateStrings)+1))
		args = append(args, *req.UserID)
	}
	if req.Currency != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("currency=$%d", len(updateStrings)+1))
		args = append(args, *req.Currency)
	}
	if req.Type != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("type=$%d", len(updateStrings)+1))
		args = append(args, *req.Type)
	}
	if req.Loaned != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("loaned=$%d", len(updateStrings)+1))
		args = append(args, *req.Loaned)
	}
	if req.Granted != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("granted=$%d", len(updateStrings)+1))
		args = append(args, *req.Granted)
	}
	if req.Repayed != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("repayed=$%d", len(updateStrings)+1))
		args = append(args, *req.Repayed)
	}
	if req.RequestID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("request_id=$%d", len(updateStrings)+1))
		args = append(args, *req.RequestID)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
