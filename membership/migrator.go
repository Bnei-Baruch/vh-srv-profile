package membership

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/keycloak"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/orders"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

type Migrator struct {
	repo          repo.ProfileRepository
	kcTokenSource keycloak.TokenSource
	ordersService orders.OrdersService
}

func NewMigrator() *Migrator {
	return new(Migrator)
}

func (m *Migrator) init() error {
	if err := m.initProfileDB(); err != nil {
		return fmt.Errorf("initProfileDB: %w", err)
	}

	m.kcTokenSource = keycloak.NewServiceClient()
	m.ordersService = orders.NewOrdersAPI()

	return nil
}

func (m *Migrator) initProfileDB() error {
	dbUrl := repo.MakeDBURL()
	db, err := repo.NewProfileDB(context.TODO(), dbUrl)
	if err != nil {
		return fmt.Errorf("repo.NewProfileDB %s: %w", dbUrl, err)
	}
	m.repo = db
	return nil
}

func (m *Migrator) migrate() error {
	users, err := m.getAllUsers()
	if err != nil {
		return fmt.Errorf("getAllUsers: %w", err)
	}
	log.Printf("Got %d users\n", len(users))

	evalResults := make(map[*uuid.UUID]repo.UserMembershipRes)
	for i, user := range users {
		if i%100 == 0 {
			log.Printf("evalUser: %d / %d\n", i, len(users))
		}
		res, err := m.evalUser(user)
		if err != nil {
			log.Printf("Error: evalUser %s: %s\n", user.UserID, err)
		} else {
			evalResults[user.UserID] = res
		}
	}
	log.Printf("%d eval results\n", len(evalResults))

	if err := m.report(users, evalResults); err != nil {
		log.Fatalf("Error generating report: %s\n", err)
	}

	return nil
}

func (m *Migrator) getAllUsers() ([]repo.User, error) {
	pageSize := 1000
	page := 0

	var allUsers []repo.User
	for {
		users, err := m.repo.GetMultipleProfiles(context.TODO(), page*pageSize, pageSize,
			"", "", "", "", "", "",
			"", "", "", "",
			"", "", "", "", "", "", "", "")
		if err != nil {
			return nil, fmt.Errorf("repo.GetMultipleProfiles: %w", err)
		}

		allUsers = append(allUsers, users...)
		page++
		if len(users) < pageSize {
			break
		}
	}

	return allUsers, nil
}

func (m *Migrator) evalUser(user repo.User) (repo.UserMembershipRes, error) {
	log.Printf("evalUser %s %s %s\n", user.UserID, user.UserInput.KeycloakID, *user.UserInput.Emails.Primary)
	ids := repo.EmailKeycloakAndUserIDBody{
		UserID:     utils.PointerString(user.UserID.String()),
		KeycloakID: utils.PointerString(user.UserInput.KeycloakID.String()),
		Email:      user.UserInput.Emails.Primary,
	}

	ctx := context.WithValue(context.Background(), common.CtxTokenSource, m.kcTokenSource)
	res, err := m.repo.EvaluateMembershipByUserID(ctx, ids)
	if err != nil {
		return repo.UserMembershipRes{}, fmt.Errorf("repo.EvaluateMembershipByUserID: %w", err)
	}

	return res, nil
}

func (m *Migrator) report(users []repo.User, evalResults map[*uuid.UUID]repo.UserMembershipRes) error {
	file, err := os.Create("eval_results.csv")
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
		"order_flag",
		"order_starting_date",
		"notification_slugs",
		"notification_count",
	}); err != nil {
		return fmt.Errorf("csv.Writer.Write header: %w", err)
	}

	ctxWithTokenSource := context.WithValue(context.Background(), common.CtxTokenSource, m.kcTokenSource)

	for i, user := range users {
		if i%100 == 0 {
			log.Printf("report: %d / %d\n", i, len(users))
		}

		evalRes, ok := evalResults[user.UserID]
		if !ok {
			log.Printf("Error user has no eval result: %s\n", user.UserID)
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

		oldStatus, err := m.ordersService.StatusByEmail(ctxWithTokenSource, *user.UserInput.Emails.Primary)
		if err != nil {
			log.Printf("Error getting old status %s: %s\n", user.UserID, err)
			vals = append(vals, "error", "error", "error", "error", "error")
		} else {
			vals = append(vals,
				strconv.FormatBool(oldStatus.Membership),
				strconv.FormatBool(*evalRes.Active != oldStatus.Membership))

			if !*evalRes.Active && !oldStatus.Membership {
				vals = append(vals, "0")
			} else if *evalRes.Active && oldStatus.Membership {
				vals = append(vals, "1")
			} else if !*evalRes.Active && oldStatus.Membership {
				vals = append(vals, "2")
			} else if *evalRes.Active && !oldStatus.Membership {
				vals = append(vals, "3")
			}

			newIsSpecial := *evalRes.Type == "special"
			vals = append(vals,
				strconv.FormatBool(oldStatus.IsSpecial),
				strconv.FormatBool(newIsSpecial != oldStatus.IsSpecial))

			if !newIsSpecial && !oldStatus.IsSpecial {
				vals = append(vals, "0")
			} else if newIsSpecial && oldStatus.IsSpecial {
				vals = append(vals, "1")
			} else if !newIsSpecial && oldStatus.IsSpecial {
				vals = append(vals, "2")
			} else if newIsSpecial && !oldStatus.IsSpecial {
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
			vals = append(vals, strconv.Itoa(*evalRes.Details.Payment.Amount))
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
		} else if *evalRes.Type == "new" || *evalRes.Type == "cancelled" {
			vals = append(vals, "", "", "", "", "")
		} else {
			log.Printf("Error unexpected membership type %s: %s\n", *evalRes.Type, user.UserID)
			vals = append(vals, "error", "error", "error", "error", "error")
		}

		if orderID > 0 {
			order, err := m.ordersService.GetOrderByID(ctxWithTokenSource, orderID)
			if err != nil {
				log.Printf("Error getting order extras %s %d: %s\n", user.UserID, *evalRes.Details.Automatic.OrderID, err)
				vals = append(vals, "error", "error")
			} else {
				vals = append(vals, order.Flag)
				if order.StartingDate.IsZero() {
					vals = append(vals, order.StartingDate.Format(time.RFC3339))
				} else {
					vals = append(vals, "")
				}
			}
		}

		notifications, err := m.repo.GetActiveUserNotificationByUserID(context.TODO(), user.UserID.String())
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
