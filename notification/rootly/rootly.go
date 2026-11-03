// Package rootly implements the notification.Service interface for Rootly
// (https://rootly.com), using its JSON:API v1 REST API.
package rootly

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/conall/outalator/notification"
)

const defaultAPIURL = "https://api.rootly.com"

// Service implements the notification.Service interface for Rootly
type Service struct {
	apiKey string
	apiURL string
	client *http.Client
}

// Config holds Rootly configuration
type Config struct {
	APIKey string
	APIURL string // Optional, defaults to the public Rootly API
}

// New creates a new Rootly notification service
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
func (s *Service) Name() string { return "rootly" }

type apiIncident struct {
	ID         string `json:"id"`
	Attributes struct {
		Title          string     `json:"title"`
		Summary        string     `json:"summary"`
		Severity       any        `json:"severity"`
		CreatedAt      time.Time  `json:"created_at"`
		StartedAt      *time.Time `json:"started_at"`
		AcknowledgedAt *time.Time `json:"acknowledged_at"`
		MitigatedAt    *time.Time `json:"mitigated_at"`
		ResolvedAt     *time.Time `json:"resolved_at"`
	} `json:"attributes"`
}

// severityName extracts a name from Rootly's severity, which may be a plain
// string or a JSON:API-style object depending on the endpoint/version.
func severityName(v any) string {
	switch sv := v.(type) {
	case string:
		return sv
	case map[string]any:
		if d, ok := sv["data"].(map[string]any); ok {
			sv = d
		}
		if a, ok := sv["attributes"].(map[string]any); ok {
			sv = a
		}
		for _, k := range []string{"name", "slug"} {
			if n, ok := sv[k].(string); ok && n != "" {
				return n
			}
		}
	}
	return ""
}

func (i apiIncident) toAlert() *notification.Alert {
	a := i.Attributes
	triggered := a.CreatedAt
	if a.StartedAt != nil && !a.StartedAt.IsZero() {
		triggered = *a.StartedAt
	}
	return &notification.Alert{
		ExternalID:     i.ID,
		Source:         "rootly",
		TeamName:       "unknown",
		Title:          a.Title,
		Description:    a.Summary,
		Severity:       severityName(a.Severity),
		TriggeredAt:    triggered,
		AcknowledgedAt: a.AcknowledgedAt,
		ResolvedAt:     a.ResolvedAt,
	}
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
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request to Rootly failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("rootly API error: %s (status: %d)", string(body), resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// FetchAlert retrieves a single incident by ID from Rootly
func (s *Service) FetchAlert(ctx context.Context, alertID string) (*notification.Alert, error) {
	var result struct {
		Data apiIncident `json:"data"`
	}
	if err := s.get(ctx, "/v1/incidents/"+url.PathEscape(alertID), nil, &result); err != nil {
		return nil, err
	}
	return result.Data.toAlert(), nil
}

// FetchRecentAlerts retrieves incidents created since the given time
func (s *Service) FetchRecentAlerts(ctx context.Context, since time.Time) ([]*notification.Alert, error) {
	var all []*notification.Alert
	const pageSize = 100
	for offset := 0; ; offset += pageSize {
		alerts, more, err := s.FetchHistoricalIncidents(ctx, HistoricalFetchOptions{
			Since: since, Limit: pageSize, Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, alerts...)
		if !more {
			return all, nil
		}
	}
}

// HistoricalFetchOptions holds options for fetching historical incidents
type HistoricalFetchOptions struct {
	Since   time.Time
	Until   time.Time
	TeamIDs []string // Rootly team slugs
	Limit   int
	Offset  int // must be a multiple of Limit; converted to a page number
}

// Team represents a team from Rootly
type Team struct {
	ID   string
	Name string
	Slug string
}

// FetchHistoricalIncidents retrieves one page of incidents with filtering.
// It returns whether more pages remain.
func (s *Service) FetchHistoricalIncidents(ctx context.Context, opts HistoricalFetchOptions) ([]*notification.Alert, bool, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	q := url.Values{}
	q.Set("page[size]", strconv.Itoa(limit))
	q.Set("page[number]", strconv.Itoa(opts.Offset/limit+1))
	q.Set("sort", "-created_at")
	if !opts.Since.IsZero() {
		q.Set("filter[created_at][gte]", opts.Since.UTC().Format(time.RFC3339))
	}
	if !opts.Until.IsZero() {
		q.Set("filter[created_at][lte]", opts.Until.UTC().Format(time.RFC3339))
	}
	if len(opts.TeamIDs) > 0 {
		q.Set("filter[teams]", strings.Join(opts.TeamIDs, ","))
	}

	var result struct {
		Data []apiIncident `json:"data"`
		Meta struct {
			NextPage *int `json:"next_page"`
		} `json:"meta"`
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	if err := s.get(ctx, "/v1/incidents", q, &result); err != nil {
		return nil, false, err
	}

	alerts := make([]*notification.Alert, 0, len(result.Data))
	for _, inc := range result.Data {
		alerts = append(alerts, inc.toAlert())
	}
	more := result.Meta.NextPage != nil || result.Links.Next != ""
	return alerts, more, nil
}

// ListTeams retrieves all teams from Rootly
func (s *Service) ListTeams(ctx context.Context) ([]Team, error) {
	var teams []Team
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page[size]", "100")
		q.Set("page[number]", strconv.Itoa(page))
		var result struct {
			Data []struct {
				ID         string `json:"id"`
				Attributes struct {
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"attributes"`
			} `json:"data"`
			Meta struct {
				NextPage *int `json:"next_page"`
			} `json:"meta"`
		}
		if err := s.get(ctx, "/v1/teams", q, &result); err != nil {
			return nil, err
		}
		for _, t := range result.Data {
			teams = append(teams, Team{ID: t.ID, Name: t.Attributes.Name, Slug: t.Attributes.Slug})
		}
		if result.Meta.NextPage == nil {
			return teams, nil
		}
	}
}

// WebhookHandler returns an HTTP handler for Rootly webhooks
func (s *Service) WebhookHandler() interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})
}
