package rootly

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
		if r.URL.Path != "/v1/incidents/abc" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("bad auth header")
		}
		_, _ = w.Write([]byte(`{"data":{"id":"abc","type":"incidents","attributes":{
			"title":"api 500s","summary":"s","severity":"sev1",
			"created_at":"2024-03-01T00:00:00Z","started_at":"2024-02-29T23:00:00Z",
			"resolved_at":"2024-03-01T02:00:00Z"}}}`))
	})
	a, err := svc.FetchAlert(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if a.Source != "rootly" || a.Title != "api 500s" || a.Severity != "sev1" || a.ResolvedAt == nil ||
		!a.TriggeredAt.Equal(time.Date(2024, 2, 29, 23, 0, 0, 0, time.UTC)) {
		t.Errorf("unexpected alert: %+v", a)
	}
}

func TestSeverityObject(t *testing.T) {
	if got := severityName(map[string]any{"data": map[string]any{"attributes": map[string]any{"name": "Critical"}}}); got != "Critical" {
		t.Errorf("got %q", got)
	}
	if got := severityName(nil); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestFetchHistoricalIncidents(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("page[size]") != "10" || q.Get("page[number]") != "3" {
			t.Errorf("paging = %q/%q", q.Get("page[size]"), q.Get("page[number]"))
		}
		if q.Get("filter[teams]") != "a,b" {
			t.Errorf("teams = %q", q.Get("filter[teams]"))
		}
		if q.Get("filter[created_at][gte]") == "" || q.Get("filter[created_at][lte]") == "" {
			t.Errorf("missing date filters: %v", q)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"title":"t","created_at":"2024-01-01T00:00:00Z"}}],"meta":{"next_page":4}}`))
	})
	got, more, err := svc.FetchHistoricalIncidents(context.Background(), HistoricalFetchOptions{
		Since: time.Now().Add(-time.Hour), Until: time.Now(), TeamIDs: []string{"a", "b"}, Limit: 10, Offset: 20,
	})
	if err != nil || len(got) != 1 || !more {
		t.Fatalf("got=%v more=%v err=%v", got, more, err)
	}
}

func TestListTeamsPaginates(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page[number]") == "1" {
			_, _ = w.Write([]byte(`{"data":[{"id":"1","attributes":{"name":"SRE","slug":"sre"}}],"meta":{"next_page":2}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"2","attributes":{"name":"DB","slug":"db"}}],"meta":{"next_page":null}}`))
	})
	teams, err := svc.ListTeams(context.Background())
	if err != nil || len(teams) != 2 || teams[1].Slug != "db" {
		t.Fatalf("teams=%+v err=%v", teams, err)
	}
}

func TestAPIError(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusUnauthorized) })
	if _, err := svc.FetchAlert(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
