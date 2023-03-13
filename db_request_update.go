package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func createOrUpdateGrant(p *profileManager, ctx context.Context, request newRequest, requestID int) error {

	patchErr := p.grant.patchGrant(ctx, request.grant, 0, requestID)

	if patchErr != nil {
		if patchErr.Error() == "not found" {

			grantBody := new(grantAndGrantMembership)
			grantBody.grant = request.grant

			_, createErr := p.grant.createGrant(ctx, *grantBody)
			if createErr != nil {
				return fmt.Errorf("problem creating grant: %w", createErr)
			}
			return nil
		}
		return fmt.Errorf("problem updating grant: %w", patchErr)
	}

	return nil
}

func (db *pgProfileDB) updateRequest(ctx context.Context, id int, request newRequest, p *profileManager) (string, error) {

	var kc_id string
	var reqType string

	toUpdate, toUpdateArgs := prepareRequestUpdate(request)

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE request SET %s WHERE id='%d' RETURNING keycloak_id, type`, toUpdate, id),
			toUpdateArgs...).
			Scan(&kc_id, &reqType); err != nil {
			return "", fmt.Errorf("problem updating event: %w", err)
		}

		if reqType == "hhmembership" {

			if request.grant.UserID == nil {
				var userID uuid.UUID
				// fetch userID via keycloak_id
				if err := db.QueryRow(ctx, `SELECT user_id FROM users WHERE keycloak_id=$1`, request.KeycloakId).Scan(&userID); err != nil {
					return "", fmt.Errorf("problem updating event: %w", err)
				}

				request.grant.UserID = &userID
			}

			grantErr := createOrUpdateGrant(p, ctx, request, id)
			if grantErr != nil {
				return "", fmt.Errorf("problem updating grant: %w", grantErr)
			}
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
