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

// GetOutage retrieves an outage by ID with all related data (alerts, notes, tags).
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("outage %s: %w", id, domain.ErrNotFound)
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

	// Eagerly load related data with three additional queries (N+1 by design,
	// consistent with the postgres backend). Use ListOutages for lightweight
	// pagination; call GetOutage only when the full record is needed.
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

// ListOutages retrieves a paginated list of outages. Returned outages contain
// only the core outage fields; related alerts, notes, and tags are not
// eagerly loaded (consistent with the postgres backend). Call GetOutage for
// a fully-populated record.
func (s *SQLiteStorage) ListOutages(ctx context.Context, limit, offset int) ([]*domain.Outage, error) {
	// SQLite treats LIMIT -1 as "no limit"; clamp to 0 so callers get an empty
	// result rather than a full table scan on a negative limit.
	if limit < 0 {
		limit = 0
	}
	if offset < 0 {
		offset = 0
	}
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
	defer func() { _ = rows.Close() }()

	var outages []*domain.Outage
	for rows.Next() {
		outage, parseErr := scanOutageRow(rows.Scan)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", parseErr)
		}
		outages = append(outages, outage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating outages: %w", err)
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
		return fmt.Errorf("outage %s: %w", outage.ID, domain.ErrNotFound)
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
		return fmt.Errorf("outage %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

// scanOutageRow populates an Outage from a single row using the provided scan
// function.
func scanOutageRow(scan scanFunc) (*domain.Outage, error) {
	outage := &domain.Outage{}
	var idStr, metadataJSON, customFieldsJSON string
	if err := scan(
		&idStr, &outage.Title, &outage.Description, &outage.Status,
		&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
		&metadataJSON, &customFieldsJSON,
	); err != nil {
		return nil, err
	}

	var parseErr error
	outage.ID, parseErr = uuid.Parse(idStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", parseErr)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &outage.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(customFieldsJSON), &outage.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	return outage, nil
}
