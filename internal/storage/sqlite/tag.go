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
	tag := &domain.Tag{}
	var idStr, outageIDStr, customFieldsJSON string
	err := s.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr, &outageIDStr, &tag.Key, &tag.Value, &tag.CreatedAt,
		&customFieldsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	tag.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tag id: %w", err)
	}
	tag.OutageID, err = uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", err)
	}

	if err := json.Unmarshal([]byte(customFieldsJSON), &tag.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	return tag, nil
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
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		var idStr, outageIDStr, customFieldsJSON string
		if err := rows.Scan(
			&idStr, &outageIDStr, &tag.Key, &tag.Value, &tag.CreatedAt,
			&customFieldsJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		tag.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag id: %w", err)
		}
		tag.OutageID, err = uuid.Parse(outageIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse outage id: %w", err)
		}

		if err := json.Unmarshal([]byte(customFieldsJSON), &tag.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
		}

		tags = append(tags, tag)
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
		return fmt.Errorf("tag not found")
	}
	return nil
}

// FindOutagesByTag finds outages that have a specific tag key-value pair.
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
