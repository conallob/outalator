//go:build sqlite

package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateOutage creates a new outage in the database.
func (s *SQLiteStorage) CreateOutage(ctx context.Context, outage *domain.Outage) error {
	metadataJSON, err := marshalJSONMap(outage.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(outage.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO outages (id, title, description, status, severity, created_at, updated_at, resolved_at, metadata, custom_fields)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query,
		outage.ID.String(), outage.Title, outage.Description, outage.Status,
		outage.Severity, outage.CreatedAt, outage.UpdatedAt, outage.ResolvedAt,
		string(metadataJSON), string(customFieldsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create outage: %w", err)
	}
	return nil
}

// GetOutage retrieves an outage by ID with all related data.
func (s *SQLiteStorage) GetOutage(ctx context.Context, id uuid.UUID) (*domain.Outage, error) {
	query := `
		SELECT id, title, description, status, severity, created_at, updated_at, resolved_at, metadata, custom_fields
		FROM outages
		WHERE id = ?
	`
	outage := &domain.Outage{}
	var idStr, metadataJSON, customFieldsJSON string
	err := s.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr, &outage.Title, &outage.Description, &outage.Status,
		&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
		&metadataJSON, &customFieldsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("outage not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
	}

	outage.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &outage.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(customFieldsJSON), &outage.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}

	alerts, err := s.ListAlertsByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load alerts: %w", err)
	}
	outage.Alerts = make([]domain.Alert, len(alerts))
	for i, a := range alerts {
		outage.Alerts[i] = *a
	}

	notes, err := s.ListNotesByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", err)
	}
	outage.Notes = make([]domain.Note, len(notes))
	for i, n := range notes {
		outage.Notes[i] = *n
	}

	tags, err := s.ListTagsByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	outage.Tags = make([]domain.Tag, len(tags))
	for i, t := range tags {
		outage.Tags[i] = *t
	}

	return outage, nil
}

// ListOutages retrieves a list of outages with pagination.
func (s *SQLiteStorage) ListOutages(ctx context.Context, limit, offset int) ([]*domain.Outage, error) {
	query := `
		SELECT id, title, description, status, severity, created_at, updated_at, resolved_at, metadata, custom_fields
		FROM outages
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list outages: %w", err)
	}
	defer rows.Close()

	var outages []*domain.Outage
	for rows.Next() {
		outage := &domain.Outage{}
		var idStr, metadataJSON, customFieldsJSON string
		if err := rows.Scan(
			&idStr, &outage.Title, &outage.Description, &outage.Status,
			&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
			&metadataJSON, &customFieldsJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}

		outage.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse outage id: %w", err)
		}

		if err := json.Unmarshal([]byte(metadataJSON), &outage.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		if err := json.Unmarshal([]byte(customFieldsJSON), &outage.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
		}

		outages = append(outages, outage)
	}

	return outages, nil
}

// UpdateOutage updates an existing outage.
func (s *SQLiteStorage) UpdateOutage(ctx context.Context, outage *domain.Outage) error {
	metadataJSON, err := marshalJSONMap(outage.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(outage.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		UPDATE outages
		SET title = ?, description = ?, status = ?, severity = ?, updated_at = ?, resolved_at = ?,
		    metadata = ?, custom_fields = ?
		WHERE id = ?
	`
	result, err := s.db.ExecContext(ctx, query,
		outage.Title, outage.Description, outage.Status,
		outage.Severity, outage.UpdatedAt, outage.ResolvedAt,
		string(metadataJSON), string(customFieldsJSON),
		outage.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update outage: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("outage not found")
	}
	return nil
}

// DeleteOutage deletes an outage by ID.
func (s *SQLiteStorage) DeleteOutage(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM outages WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete outage: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("outage not found")
	}
	return nil
}
