package membership

import (
	"context"
	"fmt"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type Invalidator struct {
	Evaluator
}

func NewInvalidator() *Invalidator {
	return new(Invalidator)
}

func (in *Invalidator) String() string {
	return "invalidator"
}

func (in *Invalidator) do() error {
	log.Printf("Invalidating memberships expiring before %s\n", time.Now().UTC())

	memberships, err := in.getExpiredMemberships()
	if err != nil {
		return fmt.Errorf("getExpiredMemberships: %w", err)
	}
	log.Printf("Got %d memberships to invalidate \n", len(memberships))

	evalResults := make(map[*uuid.UUID]repo.UserMembershipRes)
	for _, membership := range memberships {
		res, err := in.evalUserID(membership.UserID.String())
		if err != nil {
			log.Printf("ERROR: evalUser %s: %s\n", membership.UserID, err)
		} else {
			evalResults[membership.UserID] = res
		}
	}

	log.Printf("%d eval results\n", len(evalResults))

	active := 0
	inactive := 0
	for k, v := range evalResults {
		if *v.Active {
			active++
		} else {
			inactive++
			log.Printf("membership is now inactive, user_id: %s, type: %s expiry: %s\n", k, *v.Type, v.Expiry)
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
