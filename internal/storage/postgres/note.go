package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateNote creates a new note in the database
func (s *PostgresStorage) CreateNote(ctx context.Context, note *domain.Note) error {
	// Marshal JSON fields
	metadataJSON, err := json.Marshal(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := json.Marshal(note.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO notes (id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = s.db.ExecContext(ctx, query,
		note.ID, note.OutageID, note.Content, note.Format,
		note.Author, note.CreatedAt, note.UpdatedAt,
		metadataJSON, customFieldsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}
	return nil
}

// GetNote retrieves a note by ID
func (s *PostgresStorage) GetNote(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	query := `
		SELECT id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields
		FROM notes
		WHERE id = $1
	`
	note := &domain.Note{}
	var metadataJSON, customFieldsJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&note.ID, &note.OutageID, &note.Content, &note.Format,
		&note.Author, &note.CreatedAt, &note.UpdatedAt,
		&metadataJSON, &customFieldsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	// Unmarshal JSON fields
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &note.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	if len(customFieldsJSON) > 0 {
		if err := json.Unmarshal(customFieldsJSON, &note.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
		}
	}

	return note, nil
}

// ListNotesByOutage retrieves all notes for a specific outage
func (s *PostgresStorage) ListNotesByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Note, error) {
	query := `
		SELECT id, outage_id, content, format, author, created_at, updated_at, metadata, custom_fields
		FROM notes
		WHERE outage_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		note := &domain.Note{}
		var metadataJSON, customFieldsJSON []byte
		err := rows.Scan(
			&note.ID, &note.OutageID, &note.Content, &note.Format,
			&note.Author, &note.CreatedAt, &note.UpdatedAt,
			&metadataJSON, &customFieldsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		// Unmarshal JSON fields
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &note.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		if len(customFieldsJSON) > 0 {
			if err := json.Unmarshal(customFieldsJSON, &note.CustomFields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
			}
		}

		notes = append(notes, note)
	}

	return notes, nil
}

// UpdateNote updates an existing note
func (s *PostgresStorage) UpdateNote(ctx context.Context, note *domain.Note) error {
	// Marshal JSON fields
	metadataJSON, err := json.Marshal(note.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := json.Marshal(note.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		UPDATE notes
		SET content = $2, format = $3, updated_at = $4, metadata = $5, custom_fields = $6
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		note.ID, note.Content, note.Format, note.UpdatedAt,
		metadataJSON, customFieldsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("note not found")
	}

	return nil
}

// DeleteNote deletes a note by ID
func (s *PostgresStorage) DeleteNote(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notes WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("note not found")
	}

	return nil
}
