package membership

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
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
	slog.Info("invalidating memberships about to expire", slog.Time("before", time.Now().UTC()))

	memberships, err := in.getExpiredMemberships()
	if err != nil {
		return fmt.Errorf("getExpiredMemberships: %w", err)
	}
	slog.Info("memberships to invalidate", slog.Int("count", len(memberships)))

	evalResults := make(map[*uuid.UUID]repo.UserMembershipRes)
	for _, membership := range memberships {
		res, err := in.evalUserID(membership.UserID.String())
		if err != nil {
			slog.Error("evalUser", slog.String("user_id", membership.UserID.String()), slog.Any("err", err))
			sentry.CaptureException(err)
		} else {
			evalResults[membership.UserID] = res
		}
	}

	slog.Info("eval results", slog.Int("count", len(evalResults)))

	active := 0
	inactive := 0
	for k, v := range evalResults {
		if *v.Active {
			active++
		} else {
			inactive++
			slog.Info("membership is now inactive",
				slog.String("user_id", k.String()), slog.String("type", *v.Type), slog.Time("expiry", *v.Expiry))
		}
	}
	slog.Info("results", slog.Int("active", active), slog.Int("inactive", inactive))

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
