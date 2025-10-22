package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateOutage creates a new outage in the database
func (s *PostgresStorage) CreateOutage(ctx context.Context, outage *domain.Outage) error {
	query := `
		INSERT INTO outages (id, title, description, status, severity, created_at, updated_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.ExecContext(ctx, query,
		outage.ID, outage.Title, outage.Description, outage.Status,
		outage.Severity, outage.CreatedAt, outage.UpdatedAt, outage.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create outage: %w", err)
	}
	return nil
}

// GetOutage retrieves an outage by ID with all related data
func (s *PostgresStorage) GetOutage(ctx context.Context, id uuid.UUID) (*domain.Outage, error) {
	query := `
		SELECT id, title, description, status, severity, created_at, updated_at, resolved_at
		FROM outages
		WHERE id = $1
	`
	outage := &domain.Outage{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&outage.ID, &outage.Title, &outage.Description, &outage.Status,
		&outage.Severity, &outage.CreatedAt, &outage.UpdatedAt, &outage.ResolvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("outage not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get outage: %w", err)
	}

	// Load related alerts
	alerts, err := s.ListAlertsByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load alerts: %w", err)
	}
	outage.Alerts = make([]domain.Alert, len(alerts))
	for i, alert := range alerts {
		outage.Alerts[i] = *alert
	}

	// Load related notes
	notes, err := s.ListNotesByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", err)
	}
	outage.Notes = make([]domain.Note, len(notes))
	for i, note := range notes {
		outage.Notes[i] = *note
	}

	// Load related tags
	tags, err := s.ListTagsByOutage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	outage.Tags = make([]domain.Tag, len(tags))
	for i, tag := range tags {
		outage.Tags[i] = *tag
	}

	return outage, nil
}

// ListOutages retrieves a list of outages with pagination
func (s *PostgresStorage) ListOutages(ctx context.Context, limit, offset int) ([]*domain.Outage, error) {
	query := `
		SELECT id, title, description, status, severity, created_at, updated_at, resolved_at
		FROM outages
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list outages: %w", err)
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

// UpdateOutage updates an existing outage
func (s *PostgresStorage) UpdateOutage(ctx context.Context, outage *domain.Outage) error {
	query := `
		UPDATE outages
		SET title = $2, description = $3, status = $4, severity = $5, updated_at = $6, resolved_at = $7
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		outage.ID, outage.Title, outage.Description, outage.Status,
		outage.Severity, outage.UpdatedAt, outage.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update outage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("outage not found")
	}

	return nil
}

// DeleteOutage deletes an outage by ID
func (s *PostgresStorage) DeleteOutage(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM outages WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete outage: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("outage not found")
	}

	return nil
}
