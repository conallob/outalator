//go:build sqlite

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateNote creates a new note in the database.
func (s *SQLiteStorage) CreateNote(ctx context.Context, note *domain.Note) error {
	metadataJSON, err := marshalJSONMap(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(note.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO notes (id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query,
		note.ID.String(), note.OutageID.String(), note.Content, note.Format,
		note.Author, note.CreatedAt, note.UpdatedAt,
		string(metadataJSON), string(customFieldsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// GetNote retrieves a note by ID.
func (s *SQLiteStorage) GetNote(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	query := `
		SELECT id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields
		FROM notes
		WHERE id = ?
	`
	note, err := scanNoteRow(s.db.QueryRowContext(ctx, query, id.String()).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("note %s: %w", id, domain.ErrNotFound)
	}
	return note, err
}

// ListNotesByOutage retrieves all notes for a specific outage.
func (s *SQLiteStorage) ListNotesByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Note, error) {
	query := `
		SELECT id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields
		FROM notes
		WHERE outage_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		note, parseErr := scanNoteRow(rows.Scan)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to scan note: %w", parseErr)
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes: %w", err)
	}

	return notes, nil
}

// UpdateNote updates the content, format, metadata, and custom_fields of an
// existing note. Author is intentionally not updated — notes are immutable
// with respect to their author after creation.
func (s *SQLiteStorage) UpdateNote(ctx context.Context, note *domain.Note) error {
	metadataJSON, err := marshalJSONMap(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(note.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		UPDATE notes
		SET content = ?, format = ?, updated_at = ?, metadata = ?, custom_fields = ?
		WHERE id = ?
	`
	result, err := s.db.ExecContext(ctx, query,
		note.Content, note.Format, note.UpdatedAt,
		string(metadataJSON), string(customFieldsJSON),
		note.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("note %s: %w", note.ID, domain.ErrNotFound)
	}
	return nil
}

// DeleteNote deletes a note by ID.
func (s *SQLiteStorage) DeleteNote(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("note %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

// scanNoteRow populates a Note from a single row using the provided scan
// function. Returns sql.ErrNoRows when no row is found.
func scanNoteRow(scan scanFunc) (*domain.Note, error) {
	note := &domain.Note{}
	var idStr, outageIDStr, metadataJSON, customFieldsJSON string
	if err := scan(
		&idStr, &outageIDStr, &note.Content, &note.Format,
		&note.Author, &note.CreatedAt, &note.UpdatedAt,
		&metadataJSON, &customFieldsJSON,
	); err != nil {
		return nil, err
	}

	var parseErr error
	note.ID, parseErr = uuid.Parse(idStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse note id: %w", parseErr)
	}
	note.OutageID, parseErr = uuid.Parse(outageIDStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", parseErr)
	}
	if parseErr = json.Unmarshal([]byte(metadataJSON), &note.Metadata); parseErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", parseErr)
	}
	if parseErr = json.Unmarshal([]byte(customFieldsJSON), &note.CustomFields); parseErr != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", parseErr)
	}
	return note, nil
}
