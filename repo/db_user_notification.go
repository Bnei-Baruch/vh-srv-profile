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
		fmt.Println("--error-while-executing-query", err)
		return UserNotification{}, err
	}

	return r, nil
}

func (db *ProfileDB) CreateUserNotification(ctx context.Context, req UserNotification) error {
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
		return nil, fmt.Errorf("problem getting multiple user notifications: %w", err)
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
			return nil, fmt.Errorf("problem scanning rows: %w", err)
		}
		r = append(r, res)
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
		fmt.Println("--error-while-executing-query", err)
		return []UserNotification{}, err
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
			return []UserNotification{}, err
		}

		userNotiRes = append(userNotiRes, r)
	}

	return userNotiRes, nil
}

func (db *ProfileDB) PatchUserNotification(ctx context.Context, noti UserNotification, id int) (int, error) {
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

func (db *ProfileDB) SoftDeleteUserNotification(ctx context.Context, id int) error {
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

func (db *ProfileDB) updateAllUserNotificationToInactive(ctx context.Context, userID string) error {
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

func (db *ProfileDB) deactivateUserNotifications(ctx context.Context, userID string, slugs ...string) error {
	ids := make([]string, 0)
	for _, slug := range slugs {
		ids = append(ids, strconv.Itoa(*NotificationsRegistry.BySlug[slug].ID))
	}

	q := fmt.Sprintf("UPDATE user_notification SET active = false WHERE user_id = $1 AND notification_id IN (%s)", strings.Join(ids, ","))
	if _, err := db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf(" db.Exec: %w", err)
	}

	return nil
}

func prepareUserNotificationCreateQuery(req UserNotification) (string, string, []interface{}) {
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

func prepareUserNotificationUpdate(req UserNotification) (string, []interface{}) {
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
