package main

import (
	"context"
	"fmt"
	"strings"
)

func (db *pgProfileDB) getActiveUserNotificationByUserID(ctx context.Context, userID string) ([]userNotificationRes, error) {
	userNotiRes := []userNotificationRes{}

	rows, err := db.Query(ctx, `
		SELECT 
			id,
			user_id,
			notification_id,
			active,
			seen_at,
			notification.slug,
			created_at,
			updated_at,
			deleted_at
		FROM user_notification
		INNER JOIN notification ON notification.id = user_notification.notification_id
		WHERE user_id = $1 AND active = true
		`, userID)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []userNotificationRes{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r userNotificationRes
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.NotificationID,
			&r.Active,
			&r.SeenAt,
			&r.Slug,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []userNotificationRes{}, err
		}

		userNotiRes = append(userNotiRes, r)
	}

	return userNotiRes, nil
}

// add user notfications
func (db *pgProfileDB) createUserNotification(ctx context.Context, req userNotification) error {
	createString, numString, createQueryArgs := prepareUserNotificationCreateQuery(req)

	if len(createQueryArgs) != 0 {
		_, err := db.Exec(ctx, fmt.Sprintf(`INSERT INTO user_notification (%s) VALUES (%s)`, createString, numString),
			createQueryArgs...)
		if err != nil {
			return fmt.Errorf("problem creating request: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) updateAllUserNotificationToInactive(ctx context.Context, userID string) error {
	_, err := db.Exec(ctx, `
		UPDATE user_notification
		SET active = false
		WHERE user_id = $1
		`, userID)
	if err != nil {
		return fmt.Errorf("problem updating user notification: %w", err)
	}

	return nil
}

// fetch user notifiction id by slug
func (db *pgProfileDB) getNotificationBySlug(ctx context.Context, slug string) (notificationRes, error) {

	if slug == "" {
		return notificationRes{}, fmt.Errorf("invalid slug")
	}

	var r notificationRes
	err := db.QueryRow(ctx, `
		SELECT 
		id,
		slug,
		content,
		created_at,
		updated_at,
		deleted_at
		FROM notification
		WHERE slug = $1
		`, slug).Scan(
		&r.ID,
		&r.Slug,
		&r.Content,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.DeletedAt,
	)
	if err != nil {
		return notificationRes{}, err
	}

	return r, nil
}

func prepareUserNotificationCreateQuery(req userNotification) (string, string, []interface{}) {
	var createStrings []string
	var numString []string
	var args []interface{}

	if req.Active != nil {
		createStrings = append(createStrings, "active")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.Active)
	}
	if req.UserID != nil {
		createStrings = append(createStrings, "user_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.UserID)
	}
	if req.NotificationID != nil {
		createStrings = append(createStrings, "notification_id")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.NotificationID)
	}
	if req.SeenAt != nil {
		createStrings = append(createStrings, "seen_at")
		numString = append(numString, fmt.Sprintf("$%d", len(numString)+1))
		args = append(args, *req.SeenAt)
	}

	concatedCreateString := strings.Join(createStrings, ",")
	concatedNumString := strings.Join(numString, ",")

	return concatedCreateString, concatedNumString, args
}
