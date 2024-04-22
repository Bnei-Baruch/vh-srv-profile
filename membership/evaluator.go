package membership

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type Evaluator struct {
	repo          repo.ProfileRepository
	kcTokenSource keycloak.TokenSource
	ordersService orders.OrdersService
}

func NewEvalutator() *Evaluator {
	return new(Evaluator)
}

func (e *Evaluator) init() error {
	if err := e.initProfileDB(); err != nil {
		return fmt.Errorf("initProfileDB: %w", err)
	}

	e.kcTokenSource = keycloak.NewServiceClient()
	e.ordersService = orders.NewOrdersAPI()

	return nil
}

func (e *Evaluator) initProfileDB() error {
	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl)
	if err != nil {
		return fmt.Errorf("repo.NewProfileDB %s: %w", dbUrl, err)
	}
	e.repo = db
	return nil
}

func (e *Evaluator) evalUserID(userID string) (repo.UserMembershipRes, error) {
	return e.eval(repo.EmailKeycloakAndUserIDBody{UserID: utils.PointerString(userID)})
}

func (e *Evaluator) eval(ids repo.EmailKeycloakAndUserIDBody) (repo.UserMembershipRes, error) {
	ctx := context.WithValue(context.Background(), common.CtxTokenSource, e.kcTokenSource)
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())

	res, err := e.repo.EvaluateMembershipByUserID(ctx, ids)
	if err != nil {
		return repo.UserMembershipRes{}, fmt.Errorf("repo.EvaluateMembershipByUserID: %w", err)
	}

	return res, nil
}
