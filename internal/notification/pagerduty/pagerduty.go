package pagerduty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/conall/outalator/internal/notification"
)

// Service implements the notification.Service interface for PagerDuty
type Service struct {
	apiKey   string
	apiURL   string
	client   *http.Client
}

// Config holds PagerDuty configuration
type Config struct {
	APIKey string
	APIURL string // Optional, defaults to PagerDuty API
}

// New creates a new PagerDuty notification service
func New(cfg Config) *Service {
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.pagerduty.com"
	}

	return &Service{
		apiKey: cfg.APIKey,
		apiURL: cfg.APIURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the service name
func (s *Service) Name() string {
	return "pagerduty"
}

// FetchAlert retrieves a single alert/incident by ID from PagerDuty
func (s *Service) FetchAlert(ctx context.Context, alertID string) (*notification.Alert, error) {
	url := fmt.Sprintf("%s/incidents/%s", s.apiURL, alertID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", s.apiKey))
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PagerDuty API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Incident struct {
			ID             string    `json:"id"`
			Title          string    `json:"title"`
			Description    string    `json:"description"`
			Status         string    `json:"status"`
			Urgency        string    `json:"urgency"`
			CreatedAt      time.Time `json:"created_at"`
			AcknowledgedAt *time.Time `json:"acknowledged_at"`
			ResolvedAt     *time.Time `json:"resolved_at"`
			Service        struct {
				Summary string `json:"summary"`
			} `json:"service"`
			Teams []struct {
				Summary string `json:"summary"`
			} `json:"teams"`
		} `json:"incident"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	teamName := "unknown"
	if len(result.Incident.Teams) > 0 {
		teamName = result.Incident.Teams[0].Summary
	}

	return &notification.Alert{
		ExternalID:     result.Incident.ID,
		Source:         "pagerduty",
		TeamName:       teamName,
		Title:          result.Incident.Title,
		Description:    result.Incident.Description,
		Severity:       result.Incident.Urgency,
		TriggeredAt:    result.Incident.CreatedAt,
		AcknowledgedAt: result.Incident.AcknowledgedAt,
		ResolvedAt:     result.Incident.ResolvedAt,
	}, nil
}

// FetchRecentAlerts retrieves recent alerts from PagerDuty
func (s *Service) FetchRecentAlerts(ctx context.Context, since time.Time) ([]*notification.Alert, error) {
	url := fmt.Sprintf("%s/incidents?since=%s", s.apiURL, since.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", s.apiKey))
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alerts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PagerDuty API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Incidents []struct {
			ID             string    `json:"id"`
			Title          string    `json:"title"`
			Description    string    `json:"description"`
			Status         string    `json:"status"`
			Urgency        string    `json:"urgency"`
			CreatedAt      time.Time `json:"created_at"`
			AcknowledgedAt *time.Time `json:"acknowledged_at"`
			ResolvedAt     *time.Time `json:"resolved_at"`
			Service        struct {
				Summary string `json:"summary"`
			} `json:"service"`
			Teams []struct {
				Summary string `json:"summary"`
			} `json:"teams"`
		} `json:"incidents"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	alerts := make([]*notification.Alert, 0, len(result.Incidents))
	for _, incident := range result.Incidents {
		teamName := "unknown"
		if len(incident.Teams) > 0 {
			teamName = incident.Teams[0].Summary
		}

		alerts = append(alerts, &notification.Alert{
			ExternalID:     incident.ID,
			Source:         "pagerduty",
			TeamName:       teamName,
			Title:          incident.Title,
			Description:    incident.Description,
			Severity:       incident.Urgency,
			TriggeredAt:    incident.CreatedAt,
			AcknowledgedAt: incident.AcknowledgedAt,
			ResolvedAt:     incident.ResolvedAt,
		})
	}

	return alerts, nil
}

// WebhookHandler returns an HTTP handler for PagerDuty webhooks
func (s *Service) WebhookHandler() interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// PagerDuty webhook implementation
		// This would parse the webhook payload and return structured data
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})
}
