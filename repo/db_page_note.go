package repo

import (
	"context"
	"fmt"
	"time"
)

type pageNoteInterface interface {
	FetchPageNotes(ctx context.Context, pageId int, pageKeycloakId string) ([]PageNote, error)
	InsertPageNote(ctx context.Context, pageId int, pageKeycloakId string, authorKeycloackId string, note string) (int, error)
	DeletePageNote(ctx context.Context, noteId int, authorKeycloakId string) error
}

type PageNote struct {
	ID               int       `json:"id"`
	AuthorName       *string   `json:"author_name"`
	AuthorEmail      string    `json:"author_email"`
	AuthorKeycloakId string    `json:"author_keycloak_id"`
	CreatedAt        time.Time `json:"created_at"`
	Content          string    `json:"content"`
}

func (db *ProfileDB) FetchPageNotes(ctx context.Context, pageId int, pageKeycloakId string) ([]PageNote, error) {
	rows, err := db.Query(ctx, `SELECT pn.id, 
		u.first_name_vernacular || ' ' || u.last_name_vernacular as author_name, 
		u.primary_email as author_email,
		pn.author_keycloak_id,
		pn.created_at,
		pn.note as content
		FROM page_notes pn join users u ON pn.author_keycloak_id=u.keycloak_id 
		WHERE pn.page_id = $1 AND pn.page_keycloak_id = $2
		order by pn.created_at desc`, pageId, pageKeycloakId)
	if err != nil {
		return []PageNote{}, fmt.Errorf("db.Query: %w", err)
	}
	notes := []PageNote{}
	defer rows.Close()
	for rows.Next() {
		var note PageNote
		if err := rows.Scan(
			&note.ID,
			&note.AuthorName,
			&note.AuthorEmail,
			&note.AuthorKeycloakId,
			&note.CreatedAt,
			&note.Content,
		); err != nil {
			return []PageNote{}, fmt.Errorf("rows.Scan: %w", err)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return notes, fmt.Errorf("rows.Err: %w", err)
	}
	return notes, nil
}

func (db *ProfileDB) InsertPageNote(ctx context.Context, pageId int, pageKeycloakId string, authorKeycloackId string, note string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO page_notes (author_keycloak_id, page_id, page_keycloak_id, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`, authorKeycloackId, pageId, pageKeycloakId, note).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *ProfileDB) DeletePageNote(ctx context.Context, noteId int, authorKeycloakId string) error {
	res, err := db.Exec(ctx, `DELETE FROM page_notes WHERE id=$1 and author_keycloak_id=$2`, noteId, authorKeycloakId)
	if err != nil {
		return err
	}
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no notes were deleted")
	}
	return nil
}
