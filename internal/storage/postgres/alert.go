package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// CreateAlert creates a new alert in the database
func (s *PostgresStorage) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		INSERT INTO alerts (id, outage_id, external_id, source, team_name, title, description,
		                    severity, triggered_at, acknowledged_at, resolved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := s.db.ExecContext(ctx, query,
		alert.ID, alert.OutageID, alert.ExternalID, alert.Source, alert.TeamName,
		alert.Title, alert.Description, alert.Severity, alert.TriggeredAt,
		alert.AcknowledgedAt, alert.ResolvedAt, alert.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetAlert retrieves an alert by ID
func (s *PostgresStorage) GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at
		FROM alerts
		WHERE id = $1
	`
	alert := &domain.Alert{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.OutageID, &alert.ExternalID, &alert.Source, &alert.TeamName,
		&alert.Title, &alert.Description, &alert.Severity, &alert.TriggeredAt,
		&alert.AcknowledgedAt, &alert.ResolvedAt, &alert.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	return alert, nil
}

// GetAlertByExternalID retrieves an alert by its external ID and source
func (s *PostgresStorage) GetAlertByExternalID(ctx context.Context, externalID, source string) (*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at
		FROM alerts
		WHERE external_id = $1 AND source = $2
	`
	alert := &domain.Alert{}
	err := s.db.QueryRowContext(ctx, query, externalID, source).Scan(
		&alert.ID, &alert.OutageID, &alert.ExternalID, &alert.Source, &alert.TeamName,
		&alert.Title, &alert.Description, &alert.Severity, &alert.TriggeredAt,
		&alert.AcknowledgedAt, &alert.ResolvedAt, &alert.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}
	return alert, nil
}

// ListAlertsByOutage retrieves all alerts for a specific outage
func (s *PostgresStorage) ListAlertsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at
		FROM alerts
		WHERE outage_id = $1
		ORDER BY triggered_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		alert := &domain.Alert{}
		err := rows.Scan(
			&alert.ID, &alert.OutageID, &alert.ExternalID, &alert.Source, &alert.TeamName,
			&alert.Title, &alert.Description, &alert.Severity, &alert.TriggeredAt,
			&alert.AcknowledgedAt, &alert.ResolvedAt, &alert.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// UpdateAlert updates an existing alert
func (s *PostgresStorage) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		UPDATE alerts
		SET outage_id = $2, external_id = $3, source = $4, team_name = $5, title = $6,
		    description = $7, severity = $8, triggered_at = $9, acknowledged_at = $10,
		    resolved_at = $11
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		alert.ID, alert.OutageID, alert.ExternalID, alert.Source, alert.TeamName,
		alert.Title, alert.Description, alert.Severity, alert.TriggeredAt,
		alert.AcknowledgedAt, alert.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("alert not found")
	}

	return nil
}
