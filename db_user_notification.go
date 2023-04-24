package main

import (
	"context"
	"fmt"
	"strings"
)

func (db *pgProfileDB) getUserNotificationByID(ctx context.Context, id int) (userNotificationRes, error) {
	var r userNotificationRes
	err := db.QueryRow(ctx, `
		SELECT
			user_notification.id,
			user_id,
			notification_id,
			active,
			seen_at,
			notification.slug,
			user_notification.created_at,
			user_notification.updated_at,
			user_notification.deleted_at
		FROM user_notification
		INNER JOIN notification ON notification.id = user_notification.notification_id
		WHERE user_notification.id = $1 AND active = true
		`, id).Scan(
		&r.ID,
		&r.UserID,
		&r.NotificationID,
		&r.Active,
		&r.SeenAt,
		&r.Slug,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.DeletedAt,
	)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return userNotificationRes{}, err
	}

	return r, nil
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

func (db *pgProfileDB) getMultipleUserNotification(ctx context.Context, intSkip int, intLimit int) ([]userNotificationRes, error) {
	var r []userNotificationRes
	rows, err := db.Query(ctx, `
		SELECT
			user_notification.id,
			user_id,
			notification_id,
			active,
			seen_at,
			notification.slug,
			user_notification.created_at,
			user_notification.updated_at,
			user_notification.deleted_at
		FROM user_notification
		INNER JOIN notification ON notification.id = user_notification.notification_id
		WHERE active = true
		ORDER BY user_notification.id DESC
		LIMIT $1 OFFSET $2
		`, intLimit, intSkip)
	if err != nil {
		return nil, fmt.Errorf("problem getting multiple user notifications: %w", err)
	}

	for rows.Next() {
		var res userNotificationRes
		if err := rows.Scan(
			&res.ID,
			&res.UserID,
			&res.NotificationID,
			&res.Active,
			&res.SeenAt,
			&res.Slug,
			&res.CreatedAt,
			&res.UpdatedAt,
			&res.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("problem scanning rows: %w", err)
		}
		r = append(r, res)
	}

	return r, nil
}

func (db *pgProfileDB) patchUserNotification(ctx context.Context, noti userNotification, id int) (int, error) {

	var ID int

	toUpdate, toUpdateArgs := prepareUserNotificationUpdate(noti)

	if len(toUpdateArgs) != 0 {
		err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE user_notification SET %s WHERE id = $%d RETURNING id`, toUpdate, len(toUpdateArgs)+1), append(toUpdateArgs, id)...).Scan(&ID)
		if err != nil {
			return 0, fmt.Errorf("problem updating user notification: %w", err)
		}
	}

	return ID, nil
}

func (db *pgProfileDB) softDeleteUserNotification(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `
		UPDATE user_notification
		SET active = false
		WHERE id = $1
		`, id)
	if err != nil {
		return fmt.Errorf("problem updating user notification: %w", err)
	}

	return nil
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

func prepareUserNotificationUpdate(req userNotification) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Active != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("active=$%d", len(args)+1))
		args = append(args, *req.Active)
	}
	if req.UserID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("user_id=$%d", len(args)+1))
		args = append(args, *req.UserID)
	}
	if req.NotificationID != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("notification_id=$%d", len(args)+1))
		args = append(args, *req.NotificationID)
	}
	if req.SeenAt != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("seen_at=$%d", len(args)+1))
		args = append(args, *req.SeenAt)
	}

	concatedUpdateString := strings.Join(updateStrings, ",")

	return concatedUpdateString, args
}
