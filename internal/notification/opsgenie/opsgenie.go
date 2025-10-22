package opsgenie

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/conall/outalator/internal/notification"
)

// Service implements the notification.Service interface for OpsGenie
type Service struct {
	apiKey   string
	apiURL   string
	client   *http.Client
}

// Config holds OpsGenie configuration
type Config struct {
	APIKey string
	APIURL string // Optional, defaults to OpsGenie API
}

// New creates a new OpsGenie notification service
func New(cfg Config) *Service {
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.opsgenie.com"
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
	return "opsgenie"
}

// FetchAlert retrieves a single alert by ID from OpsGenie
func (s *Service) FetchAlert(ctx context.Context, alertID string) (*notification.Alert, error) {
	url := fmt.Sprintf("%s/v2/alerts/%s", s.apiURL, alertID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GenieKey %s", s.apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpsGenie API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Data struct {
			ID            string    `json:"id"`
			Message       string    `json:"message"`
			Description   string    `json:"description"`
			Status        string    `json:"status"`
			Priority      string    `json:"priority"`
			CreatedAt     time.Time `json:"createdAt"`
			AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
			ClosedAt      *time.Time `json:"closedAt,omitempty"`
			Teams         []struct {
				Name string `json:"name"`
			} `json:"teams"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	teamName := "unknown"
	if len(result.Data.Teams) > 0 {
		teamName = result.Data.Teams[0].Name
	}

	return &notification.Alert{
		ExternalID:     result.Data.ID,
		Source:         "opsgenie",
		TeamName:       teamName,
		Title:          result.Data.Message,
		Description:    result.Data.Description,
		Severity:       result.Data.Priority,
		TriggeredAt:    result.Data.CreatedAt,
		AcknowledgedAt: result.Data.AcknowledgedAt,
		ResolvedAt:     result.Data.ClosedAt,
	}, nil
}

// FetchRecentAlerts retrieves recent alerts from OpsGenie
func (s *Service) FetchRecentAlerts(ctx context.Context, since time.Time) ([]*notification.Alert, error) {
	// OpsGenie uses a query parameter for filtering by creation time
	query := fmt.Sprintf("createdAt > %d", since.Unix()*1000)
	url := fmt.Sprintf("%s/v2/alerts?query=%s&order=desc", s.apiURL, query)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("GenieKey %s", s.apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alerts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpsGenie API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID            string    `json:"id"`
			Message       string    `json:"message"`
			Description   string    `json:"description"`
			Status        string    `json:"status"`
			Priority      string    `json:"priority"`
			CreatedAt     time.Time `json:"createdAt"`
			AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
			ClosedAt      *time.Time `json:"closedAt,omitempty"`
			Teams         []struct {
				Name string `json:"name"`
			} `json:"teams"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	alerts := make([]*notification.Alert, 0, len(result.Data))
	for _, alert := range result.Data {
		teamName := "unknown"
		if len(alert.Teams) > 0 {
			teamName = alert.Teams[0].Name
		}

		alerts = append(alerts, &notification.Alert{
			ExternalID:     alert.ID,
			Source:         "opsgenie",
			TeamName:       teamName,
			Title:          alert.Message,
			Description:    alert.Description,
			Severity:       alert.Priority,
			TriggeredAt:    alert.CreatedAt,
			AcknowledgedAt: alert.AcknowledgedAt,
			ResolvedAt:     alert.ClosedAt,
		})
	}

	return alerts, nil
}

// WebhookHandler returns an HTTP handler for OpsGenie webhooks
func (s *Service) WebhookHandler() interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// OpsGenie webhook implementation
		// This would parse the webhook payload and return structured data
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})
}
