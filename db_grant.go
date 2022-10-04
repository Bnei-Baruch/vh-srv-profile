package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
)

func (db *pgProfileDB) getGrantByID(ctx context.Context, id int) (grantRes, error) {
	var grant grantRes

	if err := db.QueryRow(ctx, `
	SELECT id,
		amount,
		currency,
		type,
		loaned,
		granted,
		repayed,
		created_at,
		updated_at,
		deleted_at 
	FROM "grant" 
	WHERE id = $1`, id).Scan(
		&grant.ID,
		&grant.Amount,
		&grant.Currency,
		&grant.Type,
		&grant.Loaned,
		&grant.Granted,
		&grant.Repayed,
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

func (db *pgProfileDB) createGrant(ctx context.Context, req grantMembershipCreate) (int, error) {

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

func (db *pgProfileDB) createGrantMembership(ctx context.Context, req grantMembershipCreate) (int, error) {

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

func prepareGrantCreateQuery(req grantMembershipCreate) (string, string, []interface{}) {
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

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}

func prepareGrantMembershipCreateQuery(req grantMembershipCreate) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.GrantID != nil {
		createStrings = append(createStrings, "grant_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.GrantID)
	}
	if req.UserID != nil {
		createStrings = append(createStrings, "user_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.UserID)
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
