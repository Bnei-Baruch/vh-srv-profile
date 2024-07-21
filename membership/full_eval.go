package membership

import (
	"context"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type FullEvaluator struct {
	Migrator
}

func NewFullEvaluator() *FullEvaluator {
	return new(FullEvaluator)
}

func (e *FullEvaluator) String() string {
	return "full eval"
}

func (e *FullEvaluator) do() error {
	users, err := e.getAllUsers()
	if err != nil {
		return fmt.Errorf("getAllUsers: %w", err)
	}
	slog.Info("getAllUsers", slog.Int("count", len(users)))

	previousStatus := e.previousStatus(users)
	slog.Info("previous status", slog.Int("count", len(previousStatus)))

	evalResults := e.evalUsers(users)
	slog.Info("eval results", slog.Int("count", len(evalResults)))

	if err := e.report(users, evalResults, previousStatus); err != nil {
		utils.LogFatal("FullEvaluator.report", slog.Any("err", err))
	}

	return nil
}

func (e *FullEvaluator) previousStatus(users []repo.User) map[*uuid.UUID]repo.UserMembershipRes {
	results := make(map[*uuid.UUID]repo.UserMembershipRes)

	ctx := context.WithValue(context.Background(), common.CtxTokenSource, e.kcTokenSource)
	for i, user := range users {
		if i%100 == 0 {
			slog.Debug("previousStatus", slog.Int("position", i), slog.Int("total", len(users)))
		}

		res, err := e.repo.GetMembershipByUserID(ctx, user.UserID.String())
		if err != nil {
			slog.Error("previousStatus", slog.String("user_id", user.UserID.String()), slog.Any("err", err))
		} else {
			results[user.UserID] = res
		}
	}

	return results
}

func (e *FullEvaluator) report(users []repo.User, evalResults, previousStatus map[*uuid.UUID]repo.UserMembershipRes) error {
	file, err := os.Create("full_eval_results.csv")
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{
		"user_id",
		"keycloak_id",
		"email",
		"active",
		"type",
		"expiry",
		"old_status",
		"status_diff",
		"status_diff_code",
		"old_is_special",
		"is_special_diff",
		"is_special_diff_code",
		"payment_type",
		"payment_method",
		"payment_amount",
		"payment_currency",
		"payment_status",
		"payment_date",
		"order_id",
		"payment_id",
		"quantity",
		"category",
		"subcategory",
		"notification_slugs",
		"notification_count",
	}); err != nil {
		return fmt.Errorf("csv.Writer.Write header: %w", err)
	}

	for i, user := range users {
		if i%100 == 0 {
			slog.Debug("report", slog.Int("position", i), slog.Int("total", len(users)))
		}

		evalRes, ok := evalResults[user.UserID]
		if !ok {
			slog.Error("user has no eval result", slog.String("user_id", user.UserID.String()))
			continue
		}

		vals := []string{
			user.UserID.String(),
			user.UserInput.KeycloakID.String(),
			*user.UserInput.Emails.Primary,
			strconv.FormatBool(*evalRes.Active),
			*evalRes.Type,
		}

		if evalRes.Expiry != nil {
			vals = append(vals, evalRes.Expiry.Format(time.RFC3339))
		} else {
			vals = append(vals, "")
		}

		prevStatus, ok := previousStatus[user.UserID]
		if !ok {
			slog.Warn("user has no previous status", slog.String("user_id", user.UserID.String()))
			vals = append(vals, "error", "error", "error", "error", "error")
		} else {
			vals = append(vals,
				strconv.FormatBool(*prevStatus.Active),
				strconv.FormatBool(*evalRes.Active != *prevStatus.Active))

			if !*evalRes.Active && !*prevStatus.Active {
				vals = append(vals, "0")
			} else if *evalRes.Active && *prevStatus.Active {
				vals = append(vals, "1")
			} else if !*evalRes.Active && *prevStatus.Active {
				vals = append(vals, "2")
			} else if *evalRes.Active && !*prevStatus.Active {
				vals = append(vals, "3")
			}

			newIsSpecial := *evalRes.Type == "special"
			oldIsSpecial := *prevStatus.Type == "special"
			vals = append(vals,
				strconv.FormatBool(oldIsSpecial),
				strconv.FormatBool(newIsSpecial != oldIsSpecial))

			if !newIsSpecial && !oldIsSpecial {
				vals = append(vals, "0")
			} else if newIsSpecial && oldIsSpecial {
				vals = append(vals, "1")
			} else if !newIsSpecial && oldIsSpecial {
				vals = append(vals, "2")
			} else if newIsSpecial && !oldIsSpecial {
				vals = append(vals, "3")
			}
		}

		if evalRes.Details.Payment.PaymentType != nil {
			vals = append(vals, *evalRes.Details.Payment.PaymentType)
		} else {
			vals = append(vals, "")
		}
		if evalRes.Details.Payment.PaymentMethod != nil {
			vals = append(vals, *evalRes.Details.Payment.PaymentMethod)
		} else {
			vals = append(vals, "")
		}
		if evalRes.Details.Payment.Amount != nil {
			vals = append(vals, strconv.FormatFloat(*evalRes.Details.Payment.Amount, 'f', 2, 64))
		} else {
			vals = append(vals, "")
		}
		if evalRes.Details.Payment.Currency != nil {
			vals = append(vals, *evalRes.Details.Payment.Currency)
		} else {
			vals = append(vals, "")
		}
		if evalRes.Details.Payment.Status != nil {
			vals = append(vals, *evalRes.Details.Payment.Status)
		} else {
			vals = append(vals, "")
		}
		if evalRes.Details.Payment.Date != nil {
			vals = append(vals, evalRes.Details.Payment.Date.Format(time.RFC3339))
		} else {
			vals = append(vals, "")
		}

		var orderID int
		if *evalRes.Type == "automatic" {
			orderID = *evalRes.Details.Automatic.OrderID
			vals = append(vals,
				strconv.Itoa(orderID),
				strconv.Itoa(*evalRes.Details.Automatic.PaymentID),
				"", "", "",
			)
		} else if *evalRes.Type == "manual" {
			orderID = *evalRes.Details.Manual.OrderID
			vals = append(vals,
				strconv.Itoa(orderID),
				strconv.Itoa(*evalRes.Details.Manual.PaymentID),
				strconv.Itoa(*evalRes.Details.Manual.Quantity),
				"", "",
			)
		} else if *evalRes.Type == "special" {
			vals = append(vals,
				"", "", "",
				*evalRes.Details.Special.ApprovedBy,
				*evalRes.Details.Special.Type,
			)
		} else if *evalRes.Type == "helphaver" {
			vals = append(vals,
				"", "",
				strconv.Itoa(*evalRes.Details.HelpHaver.NbMonths),
				"", "",
			)
		} else if *evalRes.Type == "new" || *evalRes.Type == "cancelled" {
			vals = append(vals, "", "", "", "", "")
		} else {
			slog.Error("unexpected membership type", slog.String("user_id", user.UserID.String()), slog.String("type", *evalRes.Type))
			vals = append(vals, "error", "error", "error", "error", "error")
		}

		notifications, err := e.repo.GetActiveUserNotificationByUserID(context.TODO(), user.UserID.String())
		if err != nil {
			return fmt.Errorf("repo.GetMultipleUserNotification: %w", err)
		}
		slugs := make([]string, 0)
		for _, n := range notifications {
			slugs = append(slugs, *n.Slug)
		}
		vals = append(vals, strings.Join(slugs, "|"), strconv.Itoa(len(slugs)))

		if err := w.Write(vals); err != nil {
			return fmt.Errorf("csv.Writer.Write: %w", err)
		}
	}

	return nil
}
