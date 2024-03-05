package repo

import (
	"context"
	"fmt"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

var NotificationsRegistry = &notificationsRegistry{}

var allMembershipNotifications = []string{
	common.NotificationSlugMBNew,
	common.NotificationSlugMBCancelled,
	common.NotificationSlugMBExpirationNotice,
	common.NotificationSlugMBHasExpiredNotice,
	common.NotificationSlugMBProblemPreviousPayment,
}

var allHelpHaverNotifications = []string{
	common.NotificationSlugHHRequestReceived,
	common.NotificationSlugHHRequestApproved,
	common.NotificationSlugHHRequestRefused,
}

type notificationsRegistry struct {
	BySlug map[string]*Notification
}

func InitNotificationRegistry(db *ProfileDB) error {
	NotificationsRegistry.BySlug = make(map[string]*Notification)

	notifications, err := db.GetMultipleNotification(context.TODO(), 0, 200)
	if err != nil {
		return fmt.Errorf("db.GetMultipleNotification: %w", err)
	}

	for i := range notifications {
		NotificationsRegistry.BySlug[*notifications[i].Slug] = &notifications[i]
	}

	return nil
}
