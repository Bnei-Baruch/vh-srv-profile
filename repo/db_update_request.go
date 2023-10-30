package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type updateRequestStorage interface {
	UpdateRequest(ctx context.Context, id int, request NewRequest) (string, error)
}

func (db *ProfileDB) UpdateRequest(ctx context.Context, id int, request NewRequest) (string, error) {
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

			if request.Grant.UserID == nil {
				var userID uuid.UUID
				// fetch userID via keycloak_id
				if err := db.QueryRow(ctx, `SELECT user_id FROM users WHERE keycloak_id=$1`, request.KeycloakId).Scan(&userID); err != nil {
					return "", fmt.Errorf("problem updating event: %w", err)
				}

				request.Grant.UserID = &userID
			}

			grantErr := db.createOrUpdateGrantAndGrantMembership(ctx, request, id)
			if grantErr != nil {
				return "", fmt.Errorf("problem updating grant: %w", grantErr)
			}
		}

		return kc_id, nil
	} else {
		return "", fmt.Errorf("invalid values")
	}
}

func (db *ProfileDB) createOrUpdateGrantAndGrantMembership(ctx context.Context, request NewRequest, requestID int) error {
	request.Grant.Type = request.Type

	grantID, patchErr := db.PatchGrant(ctx, request.Grant, 0, requestID)

	if patchErr != nil {
		if patchErr.Error() == "not found" {

			grantBody := new(GrantAndGrantMembership)
			request.Grant.RequestID = &requestID
			grantBody.Grant = request.Grant
			grantBody.Month = request.GrantMembership.Month

			newGrantID, createErr := db.CreateGrant(ctx, *grantBody)
			if createErr != nil {
				return fmt.Errorf("problem creating grant: %w", createErr)
			}

			grantBody.GrantMembership.GrantID = &newGrantID
			grantBody.GrantMembership.Month = request.GrantMembership.Month
			grantBody.GrantMembership.MonthsLeft = request.GrantMembership.Month
			grantBody.GrantMembership.MonthsUsed = new(int)

			_, createErr = db.CreateGrantMembership(ctx, *grantBody)

			if createErr != nil {
				return fmt.Errorf("problem creating grant membership: %w", createErr)
			}

			return nil
		}
		return fmt.Errorf("problem updating grant: %w", patchErr)
	}

	if grantID != 0 {
		grantBody := new(GrantAndGrantMembership)
		grantBody.GrantMembership.GrantID = &grantID
		grantBody.GrantMembership.Month = request.GrantMembership.Month
		grantBody.GrantMembership.MonthsLeft = request.GrantMembership.Month
		grantBody.GrantMembership.MonthsUsed = new(int)

		_, patchErr := db.PatchGrantMembership(ctx, grantBody.GrantMembership, grantID)

		if patchErr != nil {
			return fmt.Errorf("problem creating grant membership: %w", patchErr)
		}

	}

	return nil
}

func prepareRequestUpdate(req NewRequest) (string, []interface{}) {
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
