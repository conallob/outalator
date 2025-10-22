package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateTag creates a new tag in the database
func (s *PostgresStorage) CreateTag(ctx context.Context, tag *domain.Tag) error {
	query := `
		INSERT INTO tags (id, outage_id, key, value, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		tag.ID, tag.OutageID, tag.Key, tag.Value, tag.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// GetTag retrieves a tag by ID
func (s *PostgresStorage) GetTag(ctx context.Context, id uuid.UUID) (*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at
		FROM tags
		WHERE id = $1
	`
	tag := &domain.Tag{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID, &tag.OutageID, &tag.Key, &tag.Value, &tag.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// ListTagsByOutage retrieves all tags for a specific outage
func (s *PostgresStorage) ListTagsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Tag, error) {
	query := `
		SELECT id, outage_id, key, value, created_at
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
		err := rows.Scan(
			&tag.ID, &tag.OutageID, &tag.Key, &tag.Value, &tag.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
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
		       o.created_at, o.updated_at, o.resolved_at
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
		err := rows.Scan(
			&outage.ID, &outage.Title, &outage.Description, &outage.Status,
			&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outage: %w", err)
		}
		outages = append(outages, outage)
	}

	return outages, nil
}
