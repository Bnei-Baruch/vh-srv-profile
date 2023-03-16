package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func createOrUpdateGrantAndGrantMembership(p *profileManager, ctx context.Context, request newRequest, requestID int) error {

	request.grant.Type = request.Type

	grantID, patchErr := p.grant.patchGrant(ctx, request.grant, 0, requestID)

	if patchErr != nil {
		if patchErr.Error() == "not found" {

			grantBody := new(grantAndGrantMembership)
			request.grant.RequestID = &requestID
			grantBody.grant = request.grant
			grantBody.Month = request.grantMembeship.Month

			newGrantID, createErr := p.grant.createGrant(ctx, *grantBody)
			if createErr != nil {
				return fmt.Errorf("problem creating grant: %w", createErr)
			}

			grantBody.grantMembeship.GrantID = &newGrantID
			grantBody.grantMembeship.Month = request.grantMembeship.Month
			grantBody.grantMembeship.MonthsLeft = request.grantMembeship.Month
			grantBody.grantMembeship.MonthsUsed = new(int)

			_, createErr = p.grant.createGrantMembership(ctx, *grantBody)

			if createErr != nil {
				return fmt.Errorf("problem creating grant membership: %w", createErr)
			}

			return nil
		}
		return fmt.Errorf("problem updating grant: %w", patchErr)
	}

	if grantID != 0 {
		grantBody := new(grantAndGrantMembership)
		grantBody.grantMembeship.GrantID = &grantID
		grantBody.grantMembeship.Month = request.grantMembeship.Month
		grantBody.grantMembeship.MonthsLeft = request.grantMembeship.Month
		grantBody.grantMembeship.MonthsUsed = new(int)

		_, patchErr := p.grant.patchGrantMembership(ctx, grantBody.grantMembeship, grantID)

		if patchErr != nil {
			return fmt.Errorf("problem creating grant membership: %w", patchErr)
		}

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

			grantErr := createOrUpdateGrantAndGrantMembership(p, ctx, request, id)
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
