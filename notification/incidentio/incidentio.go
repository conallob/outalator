// Package incidentio implements the notification.Service interface for
// incident.io (https://incident.io), using its v2 REST API.
package incidentio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/conall/outalator/notification"
)

const defaultAPIURL = "https://api.incident.io"

// Service implements the notification.Service interface for incident.io
type Service struct {
	apiKey string
	apiURL string
	client *http.Client
}

// Config holds incident.io configuration
type Config struct {
	APIKey string
	APIURL string // Optional, defaults to the public incident.io API
}

// New creates a new incident.io notification service
func New(cfg Config) *Service {
	if cfg.APIURL == "" {
		cfg.APIURL = defaultAPIURL
	}
	return &Service{
		apiKey: cfg.APIKey,
		apiURL: cfg.APIURL,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the service name
func (s *Service) Name() string { return "incidentio" }

type apiIncident struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Summary  string `json:"summary"`
	Severity *struct {
		Name string `json:"name"`
	} `json:"severity"`
	IncidentStatus *struct {
		Category string `json:"category"`
	} `json:"incident_status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i apiIncident) toAlert() *notification.Alert {
	a := &notification.Alert{
		ExternalID:  i.ID,
		Source:      "incidentio",
		TeamName:    "unknown", // incident.io has no first-class team on an incident
		Title:       i.Name,
		Description: i.Summary,
		TriggeredAt: i.CreatedAt,
	}
	if i.Severity != nil {
		a.Severity = i.Severity.Name
	}
	// incident.io has no acknowledge step. Treat an incident that has reached
	// a closed/post-incident state as resolved at its last update.
	if i.IncidentStatus != nil {
		switch i.IncidentStatus.Category {
		case "closed", "post-incident":
			t := i.UpdatedAt
			a.ResolvedAt = &t
		}
	}
	return a
}

func (s *Service) get(ctx context.Context, path string, q url.Values, out any) error {
	u := s.apiURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request to incident.io failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("incident.io API error: %s (status: %d)", string(body), resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// FetchAlert retrieves a single incident by ID from incident.io
func (s *Service) FetchAlert(ctx context.Context, alertID string) (*notification.Alert, error) {
	var result struct {
		Incident apiIncident `json:"incident"`
	}
	if err := s.get(ctx, "/v2/incidents/"+url.PathEscape(alertID), nil, &result); err != nil {
		return nil, err
	}
	return result.Incident.toAlert(), nil
}

// FetchRecentAlerts retrieves incidents created since the given time
func (s *Service) FetchRecentAlerts(ctx context.Context, since time.Time) ([]*notification.Alert, error) {
	var all []*notification.Alert
	cursor := ""
	for {
		alerts, next, err := s.FetchHistoricalIncidents(ctx, HistoricalFetchOptions{Since: since, Cursor: cursor})
		if err != nil {
			return nil, err
		}
		all = append(all, alerts...)
		if next == "" {
			return all, nil
		}
		cursor = next
	}
}

// HistoricalFetchOptions holds options for fetching historical incidents.
// incident.io paginates with an opaque cursor rather than an offset.
type HistoricalFetchOptions struct {
	Since  time.Time
	Until  time.Time
	Limit  int    // page size; defaults to 25
	Cursor string // value returned as next cursor by the previous call
}

// FetchHistoricalIncidents retrieves one page of incidents. It returns the
// cursor for the next page, or "" when there are no more results.
func (s *Service) FetchHistoricalIncidents(ctx context.Context, opts HistoricalFetchOptions) ([]*notification.Alert, string, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}
	q := url.Values{}
	q.Set("page_size", strconv.Itoa(limit))
	if opts.Cursor != "" {
		q.Set("after", opts.Cursor)
	}
	// The API filters on date granularity; exact bounds are re-applied below.
	if !opts.Since.IsZero() {
		q.Set("created_at[gte]", opts.Since.UTC().Format("2006-01-02"))
	}
	if !opts.Until.IsZero() {
		q.Set("created_at[lte]", opts.Until.UTC().Format("2006-01-02"))
	}

	var result struct {
		Incidents      []apiIncident `json:"incidents"`
		PaginationMeta struct {
			After string `json:"after"`
		} `json:"pagination_meta"`
	}
	if err := s.get(ctx, "/v2/incidents", q, &result); err != nil {
		return nil, "", err
	}

	alerts := make([]*notification.Alert, 0, len(result.Incidents))
	for _, inc := range result.Incidents {
		if !opts.Since.IsZero() && inc.CreatedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && inc.CreatedAt.After(opts.Until) {
			continue
		}
		alerts = append(alerts, inc.toAlert())
	}
	return alerts, result.PaginationMeta.After, nil
}

// WebhookHandler returns an HTTP handler for incident.io webhooks
func (s *Service) WebhookHandler() interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})
}
