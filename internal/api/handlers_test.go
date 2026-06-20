package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/conall/outalator/domain"
	"github.com/conall/outalator/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// memStorage is an in-memory storage for handler tests.
type memStorage struct {
	mu      sync.RWMutex
	outages map[uuid.UUID]*domain.Outage
	notes   map[uuid.UUID]*domain.Note
	tags    map[uuid.UUID]*domain.Tag
	alerts  map[uuid.UUID]*domain.Alert
}

func newMemStorage() *memStorage {
	return &memStorage{
		outages: make(map[uuid.UUID]*domain.Outage),
		notes:   make(map[uuid.UUID]*domain.Note),
		tags:    make(map[uuid.UUID]*domain.Tag),
		alerts:  make(map[uuid.UUID]*domain.Alert),
	}
}

func (m *memStorage) Close() error { return nil }

func (m *memStorage) CreateOutage(_ context.Context, o *domain.Outage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *o
	m.outages[o.ID] = &cp
	return nil
}

func (m *memStorage) GetOutage(_ context.Context, id uuid.UUID) (*domain.Outage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.outages[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *o
	for _, n := range m.notes {
		if n.OutageID == id {
			cp.Notes = append(cp.Notes, *n)
		}
	}
	for _, t := range m.tags {
		if t.OutageID == id {
			cp.Tags = append(cp.Tags, *t)
		}
	}
	return &cp, nil
}

func (m *memStorage) ListOutages(_ context.Context, limit, offset int) ([]*domain.Outage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]*domain.Outage, 0, len(m.outages))
	for _, o := range m.outages {
		cp := *o
		all = append(all, &cp)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *memStorage) UpdateOutage(_ context.Context, o *domain.Outage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.outages[o.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *o
	m.outages[o.ID] = &cp
	return nil
}

func (m *memStorage) DeleteOutage(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.outages, id)
	return nil
}

func (m *memStorage) CreateAlert(_ context.Context, a *domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.alerts[a.ID] = &cp
	return nil
}

func (m *memStorage) GetAlert(_ context.Context, id uuid.UUID) (*domain.Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.alerts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *memStorage) GetAlertByExternalID(_ context.Context, externalID, source string) (*domain.Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.alerts {
		if a.ExternalID == externalID && a.Source == source {
			cp := *a
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *memStorage) ListAlertsByOutage(_ context.Context, _ uuid.UUID) ([]*domain.Alert, error) {
	return nil, nil
}

func (m *memStorage) UpdateAlert(_ context.Context, _ *domain.Alert) error {
	return nil
}

func (m *memStorage) CreateNote(_ context.Context, n *domain.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *n
	m.notes[n.ID] = &cp
	return nil
}

func (m *memStorage) GetNote(_ context.Context, id uuid.UUID) (*domain.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	return &cp, nil
}

func (m *memStorage) ListNotesByOutage(_ context.Context, _ uuid.UUID) ([]*domain.Note, error) {
	return nil, nil
}

func (m *memStorage) UpdateNote(_ context.Context, _ *domain.Note) error {
	return nil
}

func (m *memStorage) DeleteNote(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *memStorage) CreateTag(_ context.Context, t *domain.Tag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.tags[t.ID] = &cp
	return nil
}

func (m *memStorage) GetTag(_ context.Context, _ uuid.UUID) (*domain.Tag, error) {
	return nil, domain.ErrNotFound
}

func (m *memStorage) ListTagsByOutage(_ context.Context, _ uuid.UUID) ([]*domain.Tag, error) {
	return nil, nil
}

func (m *memStorage) DeleteTag(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *memStorage) FindOutagesByTag(_ context.Context, key, value string) ([]*domain.Outage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[uuid.UUID]bool)
	var out []*domain.Outage
	for _, t := range m.tags {
		if t.Key == key && t.Value == value && !seen[t.OutageID] {
			seen[t.OutageID] = true
			if o, ok := m.outages[t.OutageID]; ok {
				cp := *o
				out = append(out, &cp)
			}
		}
	}
	return out, nil
}

// newTestHandler wires up handler + router for tests.
func newTestHandler() (*Handler, *mux.Router) {
	svc := service.New(newMemStorage())
	h := NewHandler(svc)
	r := mux.NewRouter()
	h.RegisterRoutes(r)
	return h, r
}

// encodeJSON encodes v into a new buffer, fataling the test on error.
func encodeJSON(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		t.Fatal(err)
	}
	return &buf
}

// decodeJSON decodes JSON from b into v, fataling the test on error.
func decodeJSON(t *testing.T, b *bytes.Buffer, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(b).Decode(v); err != nil {
		t.Fatal(err)
	}
}

func TestHealth(t *testing.T) {
	_, router := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Health status = %d, want 200", rr.Code)
	}
	var resp map[string]string
	decodeJSON(t, rr.Body, &resp)
	if resp["status"] != "healthy" {
		t.Errorf("status = %q, want healthy", resp["status"])
	}
}

func TestCreateOutage(t *testing.T) {
	_, router := newTestHandler()

	tests := []struct {
		name     string
		body     interface{}
		wantCode int
	}{
		{
			name:     "valid request",
			body:     domain.CreateOutageRequest{Title: "test", Severity: "low"},
			wantCode: http.StatusCreated,
		},
		{
			name:     "invalid json",
			body:     "not-json{{{",
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				body.WriteString(v)
			default:
				if err := json.NewEncoder(&body).Encode(v); err != nil {
					t.Fatal(err)
				}
			}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/outages", &body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantCode, rr.Body.String())
			}
		})
	}
}

func TestListOutages(t *testing.T) {
	_, router := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/outages", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	var resp map[string]interface{}
	decodeJSON(t, rr.Body, &resp)
	if _, ok := resp["outages"]; !ok {
		t.Error("response missing 'outages' key")
	}
}

func TestGetOutage(t *testing.T) {
	_, router := newTestHandler()

	// Create one first.
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/outages",
		encodeJSON(t, domain.CreateOutageRequest{Title: "test", Severity: "low"})))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", rr.Code, rr.Body.String())
	}
	var created domain.Outage
	decodeJSON(t, rr.Body, &created)

	tests := []struct {
		name     string
		id       string
		wantCode int
	}{
		{"found", created.ID.String(), http.StatusOK},
		{"not found", uuid.New().String(), http.StatusNotFound},
		{"bad id", "not-a-uuid", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/outages/"+tt.id, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantCode)
			}
		})
	}
}

func TestUpdateOutage(t *testing.T) {
	_, router := newTestHandler()

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/outages",
		encodeJSON(t, domain.CreateOutageRequest{Title: "original", Severity: "low"})))
	var created domain.Outage
	decodeJSON(t, rr.Body, &created)

	newTitle := "updated"
	updateBody, _ := json.Marshal(domain.UpdateOutageRequest{Title: &newTitle})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/outages/"+created.ID.String(), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("UpdateOutage status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	var updated domain.Outage
	decodeJSON(t, rr.Body, &updated)
	if updated.Title != newTitle {
		t.Errorf("Title = %q, want %q", updated.Title, newTitle)
	}
}

func TestAddNote_Unauthenticated(t *testing.T) {
	_, router := newTestHandler()

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/outages",
		encodeJSON(t, domain.CreateOutageRequest{Title: "test", Severity: "low"})))
	var created domain.Outage
	decodeJSON(t, rr.Body, &created)

	// No user in context → should return 401.
	noteBody, _ := json.Marshal(domain.AddNoteRequest{Content: "test note", Format: "plaintext"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/outages/"+created.ID.String()+"/notes", bytes.NewBuffer(noteBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("AddNote (no auth) status = %d, want 401", rr.Code)
	}
}

func TestAddNote_Authenticated(t *testing.T) {
	// auth.GetUserFromContext uses an unexported context key, so the cleanest way
	// to test the authenticated AddNote path without a live session store is to
	// call AddNote directly on the handler with a request whose context was set
	// by the same auth package internals. We verify AddNote via the service layer
	// instead, since the HTTP→service path is covered by other tests.
	svc := service.New(newMemStorage())
	ctx := context.Background()
	o, err := svc.CreateOutage(ctx, domain.CreateOutageRequest{Title: "x", Severity: "low"})
	if err != nil {
		t.Fatal(err)
	}
	fakeEmail := "test@example.com"
	note, err := svc.AddNote(ctx, o.ID, domain.AddNoteRequest{
		Content: "my note",
		Format:  "plaintext",
		Author:  fakeEmail,
	})
	if err != nil {
		t.Fatal(err)
	}
	if note.Author != fakeEmail {
		t.Errorf("Author = %q, want %q", note.Author, fakeEmail)
	}
}


func TestAddTag(t *testing.T) {
	_, router := newTestHandler()

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/outages",
		encodeJSON(t, domain.CreateOutageRequest{Title: "test", Severity: "low"})))
	var created domain.Outage
	decodeJSON(t, rr.Body, &created)

	tagBody, _ := json.Marshal(map[string]string{"key": "jira", "value": "OPS-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/outages/"+created.ID.String()+"/tags", bytes.NewBuffer(tagBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("AddTag status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}
}

func TestSearchByTag(t *testing.T) {
	_, router := newTestHandler()

	// Missing params.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tags/search", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("SearchByTag (missing params) status = %d, want 400", rr.Code)
	}

	// With params.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tags/search?key=jira&value=OPS-1", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("SearchByTag status = %d, want 200", rr.Code)
	}
}
