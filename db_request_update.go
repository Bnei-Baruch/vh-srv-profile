package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func (db *pgProfileDB) updateRequest(ctx context.Context, keycloakID uuid.UUID, request newRequest) error {

	toUpdate, toUpdateArgs := prepareRequestUpdate(request)

	if len(toUpdateArgs) != 0 {
		updateRes, err := db.Exec(ctx, fmt.Sprintf(`UPDATE request SET %s WHERE keycloak_id=%s`, toUpdate, keycloakID),
			toUpdateArgs...)
		if err != nil {
			return fmt.Errorf("problem updating event: %w", err)
		}

		if updateRes.RowsAffected() == 0 {
			return fmt.Errorf("not found")
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func prepareRequestUpdate(req newRequest) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.RejectionNote != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("rejection_note=$%d", len(updateStrings)+1))
		args = append(args, req.RejectionNote)
	}
	if req.RequestName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("request_name=$%d", len(updateStrings)+1))
		args = append(args, req.RejectionNote)
	}
	if req.RequestNote != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("request_note=$%d", len(updateStrings)+1))
		args = append(args, req.RejectionNote)
	}
	if req.Status != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("status=$%d", len(updateStrings)+1))
		args = append(args, req.RejectionNote)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
