package notification

import (
	"context"
	"time"
)

// Alert represents a notification alert from an oncall service
type Alert struct {
	ExternalID     string
	Source         string // pagerduty, opsgenie, etc.
	TeamName       string
	Title          string
	Description    string
	Severity       string
	TriggeredAt    time.Time
	AcknowledgedAt *time.Time
	ResolvedAt     *time.Time
}

// Service defines the interface for oncall notification services
type Service interface {
	// Name returns the service name (e.g., "pagerduty", "opsgenie")
	Name() string

	// FetchAlert retrieves a single alert by its external ID
	FetchAlert(ctx context.Context, alertID string) (*Alert, error)

	// FetchRecentAlerts retrieves recent alerts within a time window
	FetchRecentAlerts(ctx context.Context, since time.Time) ([]*Alert, error)

	// WebhookHandler returns an HTTP handler function for receiving webhooks
	// This allows each service to implement its own webhook format
	WebhookHandler() interface{}
}
