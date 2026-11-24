package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
)

type createRequestStorage interface {
	CreateRequest(ctx context.Context, request NewRequest) error
	ConcludeRequest(ctx context.Context, reqID int, conclusion RequestConclusion) error
	NotifyHHRequest(ctx context.Context, keycloakID string, slug string) error
}

// NotifyHHRequest creates an in-app HH notification for the member, deactivating any
// previous HH notification first. Used by the orders-events handler so v2 Help Haver
// requests (stored in orders) produce the same member notifications as v1.
func (db *ProfileDB) NotifyHHRequest(ctx context.Context, keycloakID string, slug string) error {
	kcID, err := uuid.FromString(keycloakID)
	if err != nil {
		return fmt.Errorf("malformed keycloak_id [%s]: %w", keycloakID, err)
	}

	user, err := db.GetProfile(ctx, kcID)
	if err != nil {
		return fmt.Errorf("db.GetProfile: %w", err)
	}

	return db.requestNotification(ctx, user, slug)
}

func (db *ProfileDB) CreateRequest(ctx context.Context, req NewRequest) error {
	req.Status = utils.PointerString(common.RequestStatusRequested)
	req.Type = utils.PointerString(common.RequestTypeHelpHaver)

	var c sqlCols
	setIf(&c, "keycloak_id", req.KeycloakId)
	setIf(&c, "rejection_note", req.RejectionNote)
	setIf(&c, "name", req.RequestName)
	setIf(&c, "request_note", req.RequestNote)
	setIf(&c, "status", req.Status)
	setIf(&c, "type", req.Type)
	setIf(&c, "months", req.Months)
	if c.empty() {
		return errors.New("invalid values")
	}

	// get user
	kcID, err := uuid.FromString(*req.KeycloakId)
	if err != nil {
		return fmt.Errorf("malformed keycloak_id [%s]: %w", *req.KeycloakId, err)
	}

	user, err := db.GetProfile(ctx, kcID)
	if err != nil {
		return fmt.Errorf("db.GetProfile: %w", err)
	}

	// delete previous non concluded requests
	_, err = db.Exec(ctx, "DELETE FROM request WHERE keycloak_id=$1 AND status=$2",
		req.KeycloakId, common.RequestStatusRequested)
	if err != nil {
		return fmt.Errorf("delete previous requests: %w", err)
	}

	// create new request
	_, err = db.Exec(ctx, `INSERT INTO request `+c.values(), c.args...)
	if err != nil {
		return fmt.Errorf("create new request: %w", err)
	}

	// create notification
	if err := db.requestNotification(ctx, user, common.NotificationSlugHHRequestReceived); err != nil {
		return fmt.Errorf("db.requestNotification: %w", err)
	}

	return nil
}

func (db *ProfileDB) ConcludeRequest(ctx context.Context, reqID int, conclusion RequestConclusion) error {
	// get request
	req, err := db.GetRequestByID(ctx, reqID)
	if err != nil {
		return fmt.Errorf("db.GetRequestByID: %w", err)
	}

	if *req.Status != common.RequestStatusRequested {
		return errors.New("request already concluded")
	}

	// get user
	kcID, err := uuid.FromString(*req.KeycloakId)
	if err != nil {
		return fmt.Errorf("malformed keycloak_id [%s]: %w", *req.KeycloakId, err)
	}

	user, err := db.GetProfile(ctx, kcID)
	if err != nil {
		return fmt.Errorf("db.GetProfile: %w", err)
	}

	// conclude request
	if conclusion.Approved {
		err = db.approveRequest(ctx, user, *req, conclusion)
		if err != nil {
			return fmt.Errorf("approveRequest: %w", err)
		}
	} else {
		err = db.rejectRequest(ctx, user, *req, conclusion)
		if err != nil {
			return fmt.Errorf("rejectRequest: %w", err)
		}
	}

	return nil
}

func (db *ProfileDB) approveRequest(ctx context.Context, user User, req NewRequest, conclusion RequestConclusion) error {
	// cancel previous grants
	_, err := db.Exec(ctx, "UPDATE \"grant\" SET cancelled_at=$1 WHERE user_id=$2 AND cancelled_at IS NULL", time.Now(), user.UserID)
	if err != nil {
		return fmt.Errorf("cancel previous grants db.Exec: %w", err)
	}

	// create new grant
	props, _ := json.Marshal(map[string]interface{}{"months": conclusion.Months})
	var grant = Grant{
		UserID:     user.UserID,
		RequestID:  req.ID,
		Type:       utils.PointerString(common.GrantTypeMembershipMonths),
		Properties: null.JSONFrom(props),
	}

	grantID, err := db.createGrant(ctx, grant)
	if err != nil {
		return fmt.Errorf("db.createGrant: %w", err)
	}
	grant.ID = &grantID

	// update request
	req.Status = utils.PointerString(common.RequestStatusApproved)
	err = db.updateRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("db.updateRequest: %w", err)
	}

	// cancel active automatic orders
	ordersService := db.ordersServiceFactory()
	userOrders, err := ordersService.GetOrders(ctx, *user.UserInput.Emails.Primary, "globalmembership", false, "desc", 200, 0)
	if err != nil {
		return fmt.Errorf("ordersService.GetOrders: %w", err)
	}
	for _, order := range userOrders {
		if order.Type == "recurring" && (order.Status == "paid" || order.Status == "nosuccess") {
			if err = ordersService.CancelOrder(ctx, order.ID); err != nil {
				return fmt.Errorf("ordersService.CancelOrder [%d]: %w", order.ID, err)
			}
		}
	}

	// eval membership
	_, err = db.EvaluateMembershipByUserID(ctx,
		EmailKeycloakAndUserIDBody{UserID: utils.PointerString(user.UserID.String())})
	if err != nil {
		return fmt.Errorf("db.EvaluateMembershipByUserID: %w", err)
	}

	// create notification
	if err := db.requestNotification(ctx, user, common.NotificationSlugHHRequestApproved); err != nil {
		return fmt.Errorf("db.requestNotification: %w", err)
	}

	return nil
}

func (db *ProfileDB) rejectRequest(ctx context.Context, user User, req NewRequest, conclusion RequestConclusion) error {
	// update request
	req.RejectionNote = conclusion.RejectionNote
	req.Status = utils.PointerString(common.RequestStatusDenied)
	err := db.updateRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("db.updateRequest: %w", err)
	}

	// create notification
	if err := db.requestNotification(ctx, user, common.NotificationSlugHHRequestRefused); err != nil {
		return fmt.Errorf("db.requestNotification: %w", err)
	}

	return nil
}

func (db *ProfileDB) updateRequest(ctx context.Context, request NewRequest) error {
	var c sqlCols
	setIf(&c, "rejection_note", request.RejectionNote)
	setIf(&c, "name", request.RequestName)
	setIf(&c, "request_note", request.RequestNote)
	setIf(&c, "status", request.Status)
	setIf(&c, "type", request.Type)
	if c.empty() {
		return errors.New("invalid values")
	}
	c.set("updated_at", time.Now())

	q := `UPDATE request SET ` + c.assignments() + ` WHERE id=` + c.bind(*request.ID)
	if _, err := db.Exec(ctx, q, c.args...); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

func (db *ProfileDB) requestNotification(ctx context.Context, user User, slug string) error {
	err := db.deactivateUserNotifications(ctx, user.UserID.String(), allHelpHaverNotifications...)
	if err != nil {
		return fmt.Errorf("db.deactivateUserNotifications: %w", err)
	}

	err = db.CreateUserNotification(ctx, UserNotification{
		UserID:         utils.PointerString(user.UserID.String()),
		NotificationID: NotificationsRegistry.BySlug[slug].ID,
		Active:         utils.PointerBool(true),
		SeenAt:         nil,
	})
	if err != nil {
		return fmt.Errorf("db.CreateUserNotification: %w", err)
	}

	return nil
}
