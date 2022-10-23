package main

import (
	"context"
	"fmt"
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
