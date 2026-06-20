package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conall/outalator/domain"
	"github.com/conall/outalator/internal/auth"
	"github.com/conall/outalator/internal/testutil"
	"github.com/conall/outalator/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// newTestHandler wires up handler + router backed by an in-memory storage.
func newTestHandler() (*Handler, *mux.Router) {
	svc := service.New(testutil.NewMemStorage())
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
	h, _ := newTestHandler()

	// Create an outage directly via the service so we have a valid ID.
	svc := service.New(testutil.NewMemStorage())
	o, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "test", Severity: "low"})
	if err != nil {
		t.Fatal(err)
	}

	noteBody, _ := json.Marshal(domain.AddNoteRequest{Content: "test note", Format: "plaintext"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/outages/"+o.ID.String()+"/notes", bytes.NewBuffer(noteBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": o.ID.String()})

	rr := httptest.NewRecorder()
	h.AddNote(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("AddNote (no auth) status = %d, want 401", rr.Code)
	}
}

func TestAddNote_Authenticated(t *testing.T) {
	mem := testutil.NewMemStorage()
	svc := service.New(mem)
	h := NewHandler(svc)

	// Create outage.
	o, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "test", Severity: "low"})
	if err != nil {
		t.Fatal(err)
	}

	noteBody, _ := json.Marshal(domain.AddNoteRequest{Content: "my note", Format: "plaintext"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/outages/"+o.ID.String()+"/notes", bytes.NewBuffer(noteBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": o.ID.String()})

	// Inject authenticated user via auth.WithUser.
	fakeUser := &auth.UserInfo{Email: "alice@example.com", Name: "Alice", Sub: "sub-123"}
	req = req.WithContext(auth.WithUser(req.Context(), fakeUser))

	rr := httptest.NewRecorder()
	h.AddNote(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("AddNote (authenticated) status = %d, want 201; body: %s", rr.Code, rr.Body.String())
	}

	var note domain.Note
	decodeJSON(t, rr.Body, &note)
	if note.Author != fakeUser.Email {
		t.Errorf("Author = %q, want %q", note.Author, fakeUser.Email)
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
