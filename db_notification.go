package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

func (db *pgProfileDB) getActiveUserNotificationByUserID(ctx context.Context, userID string) ([]userNotificationRes, error) {
	userNotiRes := []userNotificationRes{}

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

func (db *pgProfileDB) getNotificationByID(ctx context.Context, id int) (notificationRes, error) {
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
		WHERE id = $1
		`, id).Scan(
		&r.ID,
		&r.Slug,
		&r.Content,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return notificationRes{}, errNotFound
		}
		return notificationRes{}, err
	}

	return r, nil
}

func (db *pgProfileDB) createNotification(ctx context.Context, noti notification) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO notification (slug, content)
		VALUES ($1, $2)
		RETURNING id
		`, noti.Slug, noti.Content).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *pgProfileDB) getMultipleNotification(ctx context.Context, intSkip int, intLimit int) ([]notificationRes, error) {
	notifRes := []notificationRes{}

	rows, err := db.Query(ctx, `
		SELECT 
		id,
		slug,
		content,
		created_at,
		updated_at,
		deleted_at
		FROM notification
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
		`, intLimit, intSkip)
	if err != nil {
		fmt.Println("--error-while-executing-query", err)
		return []notificationRes{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var r notificationRes
		if err := rows.Scan(
			&r.ID,
			&r.Slug,
			&r.Content,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []notificationRes{}, err
		}

		notifRes = append(notifRes, r)
	}

	return notifRes, nil
}

func (db *pgProfileDB) patchNotification(ctx context.Context, noti notification, id int) (int, error) {

	var ID int

	toUpdate, toUpdateArgs := prepareNotificationUpdate(noti)

	if len(toUpdateArgs) != 0 {
		if err := db.QueryRow(ctx, fmt.Sprintf(`UPDATE notification SET %s WHERE id='%d' RETURNING id`, toUpdate, id),
			toUpdateArgs...).
			Scan(&ID); err != nil {
			return 0, fmt.Errorf("problem updating notification: %w", err)
		}

		return ID, nil
	} else {
		return 0, fmt.Errorf("invalid values")
	}
}

func (db *pgProfileDB) softDeleteNotification(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `
		UPDATE notification
		SET deleted_at = $1
		WHERE id = $2
		`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("problem deleting notification: %w", err)
	}

	return nil
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

func prepareNotificationUpdate(req notification) (string, []interface{}) {
	var updateStrings []string
	var args []interface{}

	if req.Slug != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("slug=$%d", len(updateStrings)+1))
		args = append(args, *req.Slug)
	}
	if req.Content != nil {
		updateStrings = append(updateStrings, fmt.Sprintf("content=$%d", len(updateStrings)+1))
		args = append(args, *req.Content)
	}

	if len(args) != 0 {
		updateStrings = append(updateStrings, fmt.Sprintf("updated_at=$%d", len(updateStrings)+1))
		args = append(args, time.Now())
	}

	updateArgument := strings.Join(updateStrings, ",")

	return updateArgument, args
}
