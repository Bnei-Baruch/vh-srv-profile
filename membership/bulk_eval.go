package membership

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/events"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type BulkEvaluator struct {
	Evaluator
}

func NewBulkEvaluator() *BulkEvaluator {
	return new(BulkEvaluator)
}

func (be *BulkEvaluator) String() string {
	return "bulk eval"
}

func (be *BulkEvaluator) BuildEvent(eventType string, payload map[string]interface{}) events.Event {
	event := events.MakeEvent(eventType, payload)
	event.Component = events.ComponentMembershipBulkEval
	event.Actor = events.ActorSystem
	return event
}

func (be *BulkEvaluator) do() error {
	ids, err := be.readIDs()
	if err != nil {
		return fmt.Errorf("readIDs: %w", err)
	}
	slog.Info("readIDs", slog.Int("count", len(ids)))

	ctx := context.WithValue(context.Background(), common.CtxEventBuilder, be)
	for i := range ids {
		if _, err := be.eval(ctx, *ids[i]); err != nil {
			slog.Error("evaluator.eval", slog.Int("line", i+1), slog.Any("err", err))
		}
	}

	return nil
}

func (be *BulkEvaluator) readIDs() ([]*repo.EmailKeycloakAndUserIDBody, error) {
	file, err := os.Open("bulk_eval.csv")
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	allIDs := make([]*repo.EmailKeycloakAndUserIDBody, 0)
	line := 1
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reader.Read [%d]: %w", line, err)
		}
		line++

		if len(rec) == 0 {
			continue
		}

		ids := new(repo.EmailKeycloakAndUserIDBody)
		if rec[0] != "" {
			ids.KeycloakID = utils.PointerString(rec[0])
		}
		if len(rec) > 1 && rec[1] != "" {
			ids.UserID = utils.PointerString(rec[1])
		}
		if len(rec) > 2 && rec[2] != "" {
			ids.Email = utils.PointerString(rec[2])
		}

		allIDs = append(allIDs, ids)
	}

	return allIDs, nil
}
