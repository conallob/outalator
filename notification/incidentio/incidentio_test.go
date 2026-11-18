package incidentio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestService(t *testing.T, h http.HandlerFunc) *Service {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(Config{APIKey: "k", APIURL: srv.URL})
}

func TestFetchAlert(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/incidents/INC-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Errorf("auth = %q", got)
		}
		_, _ = w.Write([]byte(`{"incident":{"id":"INC-1","name":"db down","summary":"s",
			"severity":{"name":"Major"},"incident_status":{"category":"closed"},
			"created_at":"2024-01-02T03:04:05Z","updated_at":"2024-01-02T04:00:00Z"}}`))
	})
	a, err := svc.FetchAlert(context.Background(), "INC-1")
	if err != nil {
		t.Fatal(err)
	}
	if a.Source != "incidentio" || a.Title != "db down" || a.Severity != "Major" || a.ResolvedAt == nil {
		t.Errorf("unexpected alert: %+v", a)
	}
}

func TestFetchAlertError(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	})
	if _, err := svc.FetchAlert(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchHistoricalIncidentsPagination(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("page_size") != "2" {
			t.Errorf("page_size = %q", q.Get("page_size"))
		}
		if q.Get("created_at[gte]") != "2024-01-01" {
			t.Errorf("gte = %q", q.Get("created_at[gte]"))
		}
		if q.Get("after") == "" {
			_, _ = w.Write([]byte(`{"incidents":[
				{"id":"a","name":"old","created_at":"2023-12-31T23:00:00Z"},
				{"id":"b","name":"new","created_at":"2024-01-01T10:00:00Z","incident_status":{"category":"active"}}],
				"pagination_meta":{"after":"cur1"}}`))
			return
		}
		if q.Get("after") != "cur1" {
			t.Errorf("after = %q", q.Get("after"))
		}
		_, _ = w.Write([]byte(`{"incidents":[{"id":"c","name":"newer","created_at":"2024-01-03T00:00:00Z"}],"pagination_meta":{}}`))
	})
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got, next, err := svc.FetchHistoricalIncidents(context.Background(), HistoricalFetchOptions{Since: since, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	// "old" is dropped by the exact-time client-side filter.
	if len(got) != 1 || got[0].ExternalID != "b" || got[0].ResolvedAt != nil || next != "cur1" {
		t.Fatalf("page1 = %+v next=%q", got, next)
	}
	got, next, err = svc.FetchHistoricalIncidents(context.Background(), HistoricalFetchOptions{Since: since, Limit: 2, Cursor: next})
	if err != nil || len(got) != 1 || next != "" {
		t.Fatalf("page2 = %+v next=%q err=%v", got, next, err)
	}
}

func TestFetchRecentAlertsFollowsCursor(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("after") == "" {
			_, _ = w.Write([]byte(`{"incidents":[{"id":"a","created_at":"2024-02-01T00:00:00Z"}],"pagination_meta":{"after":"n"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"incidents":[{"id":"b","created_at":"2024-02-02T00:00:00Z"}],"pagination_meta":{}}`))
	})
	got, err := svc.FetchRecentAlerts(context.Background(), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || len(got) != 2 {
		t.Fatalf("got %d alerts, err=%v", len(got), err)
	}
}
