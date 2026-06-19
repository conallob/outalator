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

// CreateTag creates a new tag in the database.
func (s *SQLiteStorage) CreateTag(ctx context.Context, tag *domain.Tag) error {
	customFieldsJSON, err := marshalJSONAny(tag.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO tags (id, outage_id, key, value, created_at, custom_fields)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query,
		tag.ID.String(), tag.OutageID.String(), tag.Key, tag.Value, tag.CreatedAt,
		string(customFieldsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// GetTag retrieves a tag by ID.
func (s *SQLiteStorage) GetTag(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at, custom_fields
		FROM tags
		WHERE id = ?
	`
	tag, err := scanTagRow(s.db.QueryRowContext(ctx, query, id.String()).Scan)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("tag %s: %w", id, domain.ErrNotFound)
	}
	return tag, err
}

// ListTagsByOutage retrieves all tags for a specific outage.
func (s *SQLiteStorage) ListTagsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at, custom_fields
		FROM tags
		WHERE outage_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []*domain.Tag
	for rows.Next() {
		tag, parseErr := scanTagRow(rows.Scan)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", parseErr)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

// DeleteTag deletes a tag by ID.
func (s *SQLiteStorage) DeleteTag(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("tag %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

// FindOutagesByTag finds outages that have a specific tag key-value pair.
// Returned outages contain only core fields; related data is not eagerly
// loaded. Call GetOutage for a fully-populated record.
func (s *SQLiteStorage) FindOutagesByTag(ctx context.Context, key, value string) ([]*domain.Outage, error) {
	query := `
		SELECT DISTINCT o.id, o.title, o.description, o.status, o.severity,
		       o.created_at, o.updated_at, o.resolved_at, o.metadata, o.custom_fields
		FROM outages o
		INNER JOIN tags t ON o.id = t.outage_id
		WHERE t.key = ? AND t.value = ?
		ORDER BY o.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to find outages by tag: %w", err)
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

// scanTagRow populates a Tag from a single row using the provided scan
// function. Returns domain.ErrNotFound when the underlying error is
// sql.ErrNoRows.
func scanTagRow(scan scanFunc) (*domain.Tag, error) {
	tag := &domain.Tag{}
	var idStr, outageIDStr, customFieldsJSON string
	if err := scan(
		&idStr, &outageIDStr, &tag.Key, &tag.Value, &tag.CreatedAt,
		&customFieldsJSON,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var parseErr error
	tag.ID, parseErr = uuid.Parse(idStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse tag id: %w", parseErr)
	}
	tag.OutageID, parseErr = uuid.Parse(outageIDStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", parseErr)
	}
	if parseErr = json.Unmarshal([]byte(customFieldsJSON), &tag.CustomFields); parseErr != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", parseErr)
	}
	return tag, nil
}
