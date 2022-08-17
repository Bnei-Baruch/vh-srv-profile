package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (db *pgProfileDB) updateRequest(ctx context.Context, id int, request newRequest) (string, error) {

	var kc_id string

	toUpdate, toUpdateArgs := prepareRequestUpdate(request)

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE request SET %s WHERE id='%d' RETURNING keycloak_id`, toUpdate, id),
			toUpdateArgs...).
			Scan(&kc_id); err != nil {
			return "", fmt.Errorf("problem updating event: %w", err)
		}

		return kc_id, nil
	} else {
		return "", fmt.Errorf("invalid values")
	}
}

func prepareRequestUpdate(req newRequest) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.RejectionNote != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("rejection_note=$%d", len(updateStrings)+1))
		args = append(args, *req.RejectionNote)
	}
	if req.RequestName != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("name=$%d", len(updateStrings)+1))
		args = append(args, *req.RequestName)
	}
	if req.RequestNote != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("request_note=$%d", len(updateStrings)+1))
		args = append(args, *req.RejectionNote)
	}
	if req.Status != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("status=$%d", len(updateStrings)+1))
		args = append(args, *req.Status)
	}
	if req.EventSlug != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("event_slug=$%d", len(updateStrings)+1))
		args = append(args, *req.EventSlug)
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
