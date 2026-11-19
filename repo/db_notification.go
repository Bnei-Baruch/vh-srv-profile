package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type notificationInterface interface {
	GetNotificationByID(ctx context.Context, id int) (Notification, error)
	CreateNotification(ctx context.Context, noti Notification) (int, error)
	GetMultipleNotification(ctx context.Context, intSkip int, intLimit int) ([]Notification, error)
	PatchNotification(ctx context.Context, noti Notification, id int) (int, error)
	SoftDeleteNotification(ctx context.Context, id int) error
}

type Notification struct {
	ID        *int       `json:"id"`
	Slug      *string    `json:"slug"`
	Content   *string    `json:"content"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (db *ProfileDB) GetNotificationByID(ctx context.Context, id int) (Notification, error) {
	var r Notification
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
			return Notification{}, common.ErrNotFound
		}
		return Notification{}, err
	}

	return r, nil
}

func (db *ProfileDB) CreateNotification(ctx context.Context, noti Notification) (int, error) {
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

func (db *ProfileDB) GetMultipleNotification(ctx context.Context, intSkip int, intLimit int) ([]Notification, error) {
	notifRes := []Notification{}

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
		return []Notification{}, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r Notification
		if err := rows.Scan(
			&r.ID,
			&r.Slug,
			&r.Content,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.DeletedAt,
		); err != nil {
			return []Notification{}, fmt.Errorf("rows.Scan: %w", err)
		}

		notifRes = append(notifRes, r)
	}

	return notifRes, nil
}

func (db *ProfileDB) PatchNotification(ctx context.Context, noti Notification, id int) (int, error) {
	var c sqlCols
	setIf(&c, "slug", noti.Slug)
	setIf(&c, "content", noti.Content)
	if c.empty() {
		return 0, fmt.Errorf("invalid values")
	}
	c.set("updated_at", time.Now())

	q := `UPDATE notification SET ` + c.assignments() + ` WHERE id=` + c.bind(id) + ` RETURNING id`
	var ID int
	if err := db.QueryRow(ctx, q, c.args...).Scan(&ID); err != nil {
		return 0, err
	}
	return ID, nil
}

func (db *ProfileDB) SoftDeleteNotification(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, `UPDATE notification SET deleted_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

// fetch user notification id by slug
func (db *ProfileDB) getNotificationBySlug(ctx context.Context, slug string) (Notification, error) {

	if slug == "" {
		return Notification{}, fmt.Errorf("invalid slug")
	}

	var r Notification
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
		return Notification{}, err
	}

	return r, nil
}
