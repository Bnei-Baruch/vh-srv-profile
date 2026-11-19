package repo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type userNotificationInterface interface {
	GetUserNotificationByID(ctx context.Context, id int) (UserNotification, error)
	CreateUserNotification(ctx context.Context, noti UserNotification) error
	GetMultipleUserNotification(ctx context.Context, intSkip int, intLimit int) ([]UserNotification, error)
	GetActiveUserNotificationByUserID(ctx context.Context, userID string) ([]UserNotification, error)
	PatchUserNotification(ctx context.Context, noti UserNotification, id int) (int, error)
	SoftDeleteUserNotification(ctx context.Context, id int) error
}

type UserNotification struct {
	ID             *int                    `json:"id"`
	UserID         *string                 `json:"user_id"`
	NotificationID *int                    `json:"notification_id"`
	Active         *bool                   `json:"active"`
	Slug           *string                 `json:"slug"`
	Content        *map[string]interface{} `json:"content"`
	SeenAt         *time.Time              `json:"seen_at"`
	CreatedAt      *time.Time              `json:"created_at"`
	UpdatedAt      *time.Time              `json:"updated_at"`
	DeletedAt      *time.Time              `json:"deleted_at"`
}

func (db *ProfileDB) GetUserNotificationByID(ctx context.Context, id int) (UserNotification, error) {
	var r UserNotification
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
		return UserNotification{}, err
	}

	return r, nil
}

func (db *ProfileDB) CreateUserNotification(ctx context.Context, req UserNotification) error {
	var c sqlCols
	setIf(&c, "active", req.Active)
	setIf(&c, "user_id", req.UserID)
	setIf(&c, "notification_id", req.NotificationID)
	setIf(&c, "seen_at", req.SeenAt)
	if c.empty() {
		return fmt.Errorf("invalid values")
	}

	_, err := db.Exec(ctx, `INSERT INTO user_notification `+c.values(), c.args...)
	return err
}

func (db *ProfileDB) GetMultipleUserNotification(ctx context.Context, intSkip int, intLimit int) ([]UserNotification, error) {
	var r []UserNotification
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
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	for rows.Next() {
		var res UserNotification
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
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		r = append(r, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return r, nil
}

func (db *ProfileDB) GetActiveUserNotificationByUserID(ctx context.Context, userID string) ([]UserNotification, error) {
	userNotiRes := []UserNotification{}

	rows, err := db.Query(ctx, `
		SELECT 
			user_notification.id,
			user_id,
			notification_id,
			active,
			seen_at,
			notification.slug,
			notification.content,
			user_notification.created_at,
			user_notification.updated_at,
			user_notification.deleted_at
		FROM user_notification
		INNER JOIN notification ON notification.id = user_notification.notification_id
		WHERE user_id = $1 AND active = true
		ORDER by user_notification.updated_at desc, user_notification.created_at desc
		`, userID)
	if err != nil {
		return []UserNotification{}, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r UserNotification
		if err := rows.Scan(
			&r.ID,
			&r.UserID,
			&r.NotificationID,
			&r.Active,
			&r.SeenAt,
			&r.Slug,
			&r.Content,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []UserNotification{}, fmt.Errorf("rows.Scan: %w", err)
		}

		userNotiRes = append(userNotiRes, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return userNotiRes, nil
}

func (db *ProfileDB) PatchUserNotification(ctx context.Context, noti UserNotification, id int) (int, error) {
	var c sqlCols
	setIf(&c, "active", noti.Active)
	setIf(&c, "user_id", noti.UserID)
	setIf(&c, "notification_id", noti.NotificationID)
	setIf(&c, "seen_at", noti.SeenAt)
	if c.empty() {
		// Nothing to update is not an error here.
		return 0, nil
	}

	q := `UPDATE user_notification SET ` + c.assignments() + ` WHERE id = ` + c.bind(id) + ` RETURNING id`
	var ID int
	if err := db.QueryRow(ctx, q, c.args...).Scan(&ID); err != nil {
		return 0, err
	}
	return ID, nil
}

func (db *ProfileDB) SoftDeleteUserNotification(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `UPDATE user_notification SET active = false WHERE id = $1`, id)
	return err
}

func (db *ProfileDB) updateAllUserNotificationToInactive(ctx context.Context, userID string) error {
	_, err := db.Exec(ctx, `UPDATE user_notification SET active = false WHERE user_id = $1`, userID)
	return err
}

func (db *ProfileDB) deactivateUserNotifications(ctx context.Context, userID string, slugs ...string) error {
	ids := make([]string, 0)
	for _, slug := range slugs {
		ids = append(ids, strconv.Itoa(*NotificationsRegistry.BySlug[slug].ID))
	}

	q := fmt.Sprintf("UPDATE user_notification SET active = false WHERE user_id = $1 AND notification_id IN (%s)",
		strings.Join(ids, ","))
	_, err := db.Exec(ctx, q, userID)
	return err
}
