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

// CreateAlert creates a new alert in the database.
func (s *SQLiteStorage) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	sourceMetadataJSON, err := marshalJSONAny(alert.SourceMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal source_metadata: %w", err)
	}
	metadataJSON, err := marshalJSONMap(alert.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(alert.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		INSERT INTO alerts (id, outage_id, external_id, source, team_name, title, description,
		                    severity, triggered_at, acknowledged_at, resolved_at, created_at,
		                    source_metadata, metadata, custom_fields)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query,
		alert.ID.String(), alert.OutageID.String(), alert.ExternalID, alert.Source, alert.TeamName,
		alert.Title, alert.Description, alert.Severity, alert.TriggeredAt,
		alert.AcknowledgedAt, alert.ResolvedAt, alert.CreatedAt,
		string(sourceMetadataJSON), string(metadataJSON), string(customFieldsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetAlert retrieves an alert by ID.
func (s *SQLiteStorage) GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at,
		       source_metadata, metadata, custom_fields
		FROM alerts
		WHERE id = ?
	`
	return s.scanAlert(s.db.QueryRowContext(ctx, query, id.String()))
}

// GetAlertByExternalID retrieves an alert by its external ID and source.
func (s *SQLiteStorage) GetAlertByExternalID(ctx context.Context, externalID, source string) (*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at,
		       source_metadata, metadata, custom_fields
		FROM alerts
		WHERE external_id = ? AND source = ?
	`
	return s.scanAlert(s.db.QueryRowContext(ctx, query, externalID, source))
}

// ListAlertsByOutage retrieves all alerts for a specific outage.
func (s *SQLiteStorage) ListAlertsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Alert, error) {
	query := `
		SELECT id, outage_id, external_id, source, team_name, title, description,
		       severity, triggered_at, acknowledged_at, resolved_at, created_at,
		       source_metadata, metadata, custom_fields
		FROM alerts
		WHERE outage_id = ?
		ORDER BY triggered_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, outageID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		alert, err := s.scanAlertRow(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

// UpdateAlert updates an existing alert.
func (s *SQLiteStorage) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	sourceMetadataJSON, err := marshalJSONAny(alert.SourceMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal source_metadata: %w", err)
	}
	metadataJSON, err := marshalJSONMap(alert.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	customFieldsJSON, err := marshalJSONAny(alert.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom_fields: %w", err)
	}

	query := `
		UPDATE alerts
		SET outage_id = ?, external_id = ?, source = ?, team_name = ?, title = ?,
		    description = ?, severity = ?, triggered_at = ?, acknowledged_at = ?,
		    resolved_at = ?, source_metadata = ?, metadata = ?, custom_fields = ?
		WHERE id = ?
	`
	result, err := s.db.ExecContext(ctx, query,
		alert.OutageID.String(), alert.ExternalID, alert.Source, alert.TeamName,
		alert.Title, alert.Description, alert.Severity, alert.TriggeredAt,
		alert.AcknowledgedAt, alert.ResolvedAt,
		string(sourceMetadataJSON), string(metadataJSON), string(customFieldsJSON),
		alert.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("alert not found")
	}
	return nil
}

// scanAlert reads one alert row from a QueryRow result.
func (s *SQLiteStorage) scanAlert(row *sql.Row) (*domain.Alert, error) {
	alert := &domain.Alert{}
	var idStr, outageIDStr, sourceMetadataJSON, metadataJSON, customFieldsJSON string
	err := row.Scan(
		&idStr, &outageIDStr, &alert.ExternalID, &alert.Source, &alert.TeamName,
		&alert.Title, &alert.Description, &alert.Severity, &alert.TriggeredAt,
		&alert.AcknowledgedAt, &alert.ResolvedAt, &alert.CreatedAt,
		&sourceMetadataJSON, &metadataJSON, &customFieldsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	alert.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alert id: %w", err)
	}
	alert.OutageID, err = uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", err)
	}

	if err := json.Unmarshal([]byte(sourceMetadataJSON), &alert.SourceMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal source_metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &alert.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(customFieldsJSON), &alert.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	return alert, nil
}

// scanAlertRow reads one alert row from a *sql.Rows cursor.
func (s *SQLiteStorage) scanAlertRow(rows *sql.Rows) (*domain.Alert, error) {
	alert := &domain.Alert{}
	var idStr, outageIDStr, sourceMetadataJSON, metadataJSON, customFieldsJSON string
	var err error
	if err = rows.Scan(
		&idStr, &outageIDStr, &alert.ExternalID, &alert.Source, &alert.TeamName,
		&alert.Title, &alert.Description, &alert.Severity, &alert.TriggeredAt,
		&alert.AcknowledgedAt, &alert.ResolvedAt, &alert.CreatedAt,
		&sourceMetadataJSON, &metadataJSON, &customFieldsJSON,
	); err != nil {
		return nil, err
	}

	alert.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alert id: %w", err)
	}
	alert.OutageID, err = uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse outage id: %w", err)
	}

	if err := json.Unmarshal([]byte(sourceMetadataJSON), &alert.SourceMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal source_metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &alert.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	if err := json.Unmarshal([]byte(customFieldsJSON), &alert.CustomFields); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom_fields: %w", err)
	}
	return alert, nil
}
