package membership

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type Evaluator struct {
	eventEmitter  events.EventEmitter
	repo          repo.ProfileRepository
	kcTokenSource keycloak.TokenSource
	ordersService orders.OrdersService
}

func NewEvalutator() *Evaluator {
	return new(Evaluator)
}

func (e *Evaluator) init() error {
	if err := e.initEventEmitter(); err != nil {
		return fmt.Errorf("initEventEmitter: %w", err)
	}

	if err := e.initProfileDB(); err != nil {
		return fmt.Errorf("initProfileDB: %w", err)
	}

	e.kcTokenSource = keycloak.NewServiceClient()
	e.ordersService = orders.NewOrdersAPI()

	return nil
}

func (e *Evaluator) initEventEmitter() error {
	var err error
	e.eventEmitter, err = events.CreateEmitter()
	if err != nil {
		return fmt.Errorf("events.CreateEmitter: %w", err)
	}
	return nil
}

func (e *Evaluator) initProfileDB() error {
	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl, e.eventEmitter)
	if err != nil {
		return fmt.Errorf("repo.NewProfileDB %s: %w", dbUrl, err)
	}
	e.repo = db
	return nil
}

func (e *Evaluator) close() {
	e.repo.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	e.eventEmitter.Close(ctx)
	cancel()
	sentry.Flush(2 * time.Second)
}

func (e *Evaluator) evalUserID(ctx context.Context, userID string) (repo.UserMembershipRes, error) {
	return e.eval(ctx, repo.EmailKeycloakAndUserIDBody{UserID: utils.PointerString(userID)})
}

func (e *Evaluator) eval(ctx context.Context, ids repo.EmailKeycloakAndUserIDBody) (repo.UserMembershipRes, error) {
	ctx2 := context.WithValue(ctx, common.CtxTokenSource, e.kcTokenSource)
	ctx2 = sentry.SetHubOnContext(ctx2, sentry.CurrentHub())

	res, err := e.repo.EvaluateMembershipByUserID(ctx2, ids)
	if err != nil {
		return repo.UserMembershipRes{}, fmt.Errorf("repo.EvaluateMembershipByUserID: %w", err)
	}

	return res, nil
}
