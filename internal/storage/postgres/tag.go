package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateTag creates a new tag in the database
func (s *PostgresStorage) CreateTag(ctx context.Context, tag *domain.Tag) error {
	// Marshal JSON fields
	customFieldsJSON, err := marshalJSONAny(tag.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO tags (id, outage_id, key, value, created_at, custom_fields)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = s.db.ExecContext(ctx, query,
		tag.ID, tag.OutageID, tag.Key, tag.Value, tag.CreatedAt,
		customFieldsJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// GetTag retrieves a tag by ID
func (s *PostgresStorage) GetTag(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at, custom_fields
		FROM tags
		WHERE id = $1
	`
	tag := &domain.Tag{}
	var customFieldsJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID, &tag.OutageID, &tag.Key, &tag.Value, &tag.CreatedAt,
		&customFieldsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	// Unmarshal JSON fields
	if len(customFieldsJSON) > 0 {
		if err := json.Unmarshal(customFieldsJSON, &tag.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
		}
	}

	return tag, nil
}

// ListTagsByOutage retrieves all tags for a specific outage
func (s *PostgresStorage) ListTagsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at, custom_fields
		FROM tags
		WHERE outage_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		var customFieldsJSON []byte
		err := rows.Scan(
			&tag.ID, &tag.OutageID, &tag.Key, &tag.Value, &tag.CreatedAt,
			&customFieldsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}

		// Unmarshal JSON fields
		if len(customFieldsJSON) > 0 {
			if err := json.Unmarshal(customFieldsJSON, &tag.CustomFields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
			}
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// DeleteTag deletes a tag by ID
func (s *PostgresStorage) DeleteTag(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// FindOutagesByTag finds outages that have a specific tag key-value pair
func (s *PostgresStorage) FindOutagesByTag(ctx context.Context, key, value string) ([]*domain.Outage, error) {
	query := `
		SELECT DISTINCT o.id, o.title, o.description, o.status, o.severity,
		       o.created_at, o.updated_at, o.resolved_at, o.metadata, o.custom_fields
		FROM outages o
		INNER JOIN tags t ON o.id = t.outage_id
		WHERE t.key = $1 AND t.value = $2
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
		var metadataJSON, customFieldsJSON []byte
		err := rows.Scan(
			&outage.ID, &outage.Title, &outage.Description, &outage.Status,
			&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
			&metadataJSON, &customFieldsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}

		// Unmarshal JSON fields
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &outage.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		if len(customFieldsJSON) > 0 {
			if err := json.Unmarshal(customFieldsJSON, &outage.CustomFields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
			}
		}

		outages = append(outages, outage)
	}

	return outages, nil
}
