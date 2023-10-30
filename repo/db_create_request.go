package repo

import (
	"context"
	"fmt"
	"strings"
)

type createRequestStorage interface {
	CreateRequest(ctx context.Context, request NewRequest) error
}

func (db *ProfileDB) CreateRequest(ctx context.Context, req NewRequest) error {

	createString, numString, createQueryArgs := prepareRequestCreateQuery(req)

	var ID int

	if len(createQueryArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`INSERT INTO request (%s) VALUES (%s) RETURNING id`, createString, numString),
			createQueryArgs...).Scan(&ID); err != nil {
			return fmt.Errorf("problem creating request: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func prepareRequestCreateQuery(req NewRequest) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.KeycloakId != nil {
		createStrings = append(createStrings, "keycloak_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.KeycloakId)
	}
	if req.RejectionNote != nil {
		createStrings = append(createStrings, "rejection_note")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RejectionNote)
	}
	if req.RequestName != nil {
		createStrings = append(createStrings, "name")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestName)
	}
	if req.RequestNote != nil {
		createStrings = append(createStrings, "request_note")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.RequestNote)
	}
	if req.Status != nil {
		createStrings = append(createStrings, "status")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Status)
	}
	if req.EventSlug != nil {
		createStrings = append(createStrings, "event_slug")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.EventSlug)
	}
	if req.Type != nil {
		createStrings = append(createStrings, "type")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Type)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
