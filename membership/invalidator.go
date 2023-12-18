package membership

import (
	"context"
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type Invalidator struct {
	repo          repo.ProfileRepository
	kcTokenSource keycloak.TokenSource
	ordersService orders.OrdersService
}

func NewInvalidator() *Invalidator {
	return new(Invalidator)
}

func (in *Invalidator) init() error {
	if err := in.initProfileDB(); err != nil {
		return fmt.Errorf("initProfileDB: %w", err)
	}

	in.kcTokenSource = keycloak.NewServiceClient()
	in.ordersService = orders.NewOrdersAPI()

	return nil
}

func (in *Invalidator) initProfileDB() error {
	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl)
	if err != nil {
		return fmt.Errorf("repo.NewProfileDB %s: %w", dbUrl, err)
	}
	in.repo = db
	return nil
}

func (in *Invalidator) invalidate() error {
	log.Printf("Invalidating memberships expiring before %s\n", time.Now().UTC())

	memberships, err := in.getExpiredMemberships()
	if err != nil {
		return fmt.Errorf("getExpiredMemberships: %w", err)
	}
	log.Printf("Got %d memberships to invalidate \n", len(memberships))

	evalResults := make(map[*uuid.UUID]repo.UserMembershipRes)
	for _, membership := range memberships {
		res, err := in.evalUser(membership)
		if err != nil {
			log.Printf("ERROR: evalUser %s: %s\n", membership.UserID, err)
		} else {
			evalResults[membership.UserID] = res
		}
	}

	log.Printf("%d eval results\n", len(evalResults))

	active := 0
	inactive := 0
	for _, v := range evalResults {
		if *v.Active {
			active++
		} else {
			inactive++
		}
	}
	log.Printf("%d active, %d inactive\n", active, inactive)

	return nil
}

func (in *Invalidator) getExpiredMemberships() ([]repo.Membership, error) {
	pageSize := 1000
	page := 0

	var expiredMemberships []repo.Membership
	for {
		memberships, err := in.repo.GetExpiredMemberships(context.TODO(), page*pageSize, pageSize)
		if err != nil {
			return nil, fmt.Errorf("repo.GetExpiredMemberships: %w", err)
		}

		expiredMemberships = append(expiredMemberships, memberships...)
		page++
		if len(memberships) < pageSize {
			break
		}
	}

	return expiredMemberships, nil
}

func (in *Invalidator) evalUser(membership repo.Membership) (repo.UserMembershipRes, error) {
	log.Printf("evalUser %s, type: %s expiry: %s\n", membership.UserID, *membership.Type, membership.Expiry)
	ids := repo.EmailKeycloakAndUserIDBody{UserID: utils.PointerString(membership.UserID.String())}

	ctx := context.WithValue(context.Background(), common.CtxTokenSource, in.kcTokenSource)
	res, err := in.repo.EvaluateMembershipByUserID(ctx, ids)
	if err != nil {
		return repo.UserMembershipRes{}, fmt.Errorf("repo.EvaluateMembershipByUserID: %w", err)
	}

	return res, nil
}
