package domain

import (
	"time"

	"github.com/google/uuid"
)

// Outage represents a tracked incident/outage created from one or more alerts
type Outage struct {
	ID           uuid.UUID         `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       string            `json:"status"` // open, investigating, resolved, closed
	Severity     string            `json:"severity"` // critical, high, medium, low
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ResolvedAt   *time.Time        `json:"resolved_at,omitempty"`
	Alerts       []Alert           `json:"alerts,omitempty"`
	Notes        []Note            `json:"notes,omitempty"`
	Tags         []Tag             `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`      // Simple key-value pairs
	CustomFields map[string]any    `json:"custom_fields,omitempty"` // Complex structured data
}

// Alert represents a paging alert from an oncall notification service
type Alert struct {
	ID               uuid.UUID         `json:"id"`
	OutageID         uuid.UUID         `json:"outage_id"`
	ExternalID       string            `json:"external_id"` // ID from PagerDuty/OpsGenie
	Source           string            `json:"source"`      // pagerduty, opsgenie, etc.
	TeamName         string            `json:"team_name"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Severity         string            `json:"severity"`
	TriggeredAt      time.Time         `json:"triggered_at"`
	AcknowledgedAt   *time.Time        `json:"acknowledged_at,omitempty"`
	ResolvedAt       *time.Time        `json:"resolved_at,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	SourceMetadata   map[string]any    `json:"source_metadata,omitempty"` // Source-specific data (PagerDuty, OpsGenie, etc.)
	Metadata         map[string]string `json:"metadata,omitempty"`        // Simple key-value pairs
	CustomFields     map[string]any    `json:"custom_fields,omitempty"`   // Complex structured data
}

// Note represents a free-form text or markdown note attached to an outage
type Note struct {
	ID           uuid.UUID         `json:"id"`
	OutageID     uuid.UUID         `json:"outage_id"`
	Content      string            `json:"content"`
	Format       string            `json:"format"` // plaintext, markdown
	Author       string            `json:"author"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`      // Simple key-value pairs
	CustomFields map[string]any    `json:"custom_fields,omitempty"` // Complex structured data
}

// Tag represents metadata attached to an outage (e.g., Jira tickets)
type Tag struct {
	ID           uuid.UUID      `json:"id"`
	OutageID     uuid.UUID      `json:"outage_id"`
	Key          string         `json:"key"`   // e.g., "jira", "service", "region"
	Value        string         `json:"value"` // e.g., "PROJ-123", "api", "us-west-2"
	CreatedAt    time.Time      `json:"created_at"`
	CustomFields map[string]any `json:"custom_fields,omitempty"` // Additional structured data
}

// CreateOutageRequest represents the data needed to create a new outage
type CreateOutageRequest struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Severity     string            `json:"severity"`
	AlertIDs     []string          `json:"alert_ids"` // External alert IDs to associate
	Tags         []struct {
		Key          string         `json:"key"`
		Value        string         `json:"value"`
		CustomFields map[string]any `json:"custom_fields,omitempty"`
	} `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CustomFields map[string]any    `json:"custom_fields,omitempty"`
}

// AddNoteRequest represents the data needed to add a note to an outage
type AddNoteRequest struct {
	Content      string            `json:"content"`
	Format       string            `json:"format"` // plaintext, markdown
	Author       string            `json:"author"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CustomFields map[string]any    `json:"custom_fields,omitempty"`
}

// UpdateOutageRequest represents the data that can be updated on an outage
type UpdateOutageRequest struct {
	Title        *string           `json:"title,omitempty"`
	Description  *string           `json:"description,omitempty"`
	Status       *string           `json:"status,omitempty"`
	Severity     *string           `json:"severity,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CustomFields map[string]any    `json:"custom_fields,omitempty"`
}
