package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/conall/outalator/domain"
	"github.com/google/uuid"
)

// memStorage is an in-memory implementation of storage.Storage for testing.
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
	// Attach related entities
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
	for _, a := range m.alerts {
		if a.OutageID == id {
			cp.Alerts = append(cp.Alerts, *a)
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

func (m *memStorage) ListAlertsByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*domain.Alert
	for _, a := range m.alerts {
		if a.OutageID == outageID {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memStorage) UpdateAlert(_ context.Context, a *domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.alerts[a.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *a
	m.alerts[a.ID] = &cp
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

func (m *memStorage) ListNotesByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*domain.Note
	for _, n := range m.notes {
		if n.OutageID == outageID {
			cp := *n
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memStorage) UpdateNote(_ context.Context, n *domain.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[n.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *n
	m.notes[n.ID] = &cp
	return nil
}

func (m *memStorage) DeleteNote(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.notes, id)
	return nil
}

func (m *memStorage) CreateTag(_ context.Context, t *domain.Tag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.tags[t.ID] = &cp
	return nil
}

func (m *memStorage) GetTag(_ context.Context, id uuid.UUID) (*domain.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tags[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *memStorage) ListTagsByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*domain.Tag
	for _, t := range m.tags {
		if t.OutageID == outageID {
			cp := *t
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memStorage) DeleteTag(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tags, id)
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

// ---- Tests ----

func TestCreateOutage(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateOutageRequest
		wantErr bool
	}{
		{
			name: "basic creation",
			req: domain.CreateOutageRequest{
				Title:       "DB is down",
				Description: "PostgreSQL primary is unresponsive",
				Severity:    "critical",
			},
		},
		{
			name: "with tags",
			req: domain.CreateOutageRequest{
				Title:    "Slow API",
				Severity: "high",
				Tags:     []domain.TagInput{{Key: "jira", Value: "OPS-123"}},
			},
		},
		{
			name: "with metadata",
			req: domain.CreateOutageRequest{
				Title:    "Disk full",
				Severity: "medium",
				Metadata: map[string]string{"region": "us-east-1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(newMemStorage())
			got, err := svc.CreateOutage(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateOutage() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if got.Title != tt.req.Title {
					t.Errorf("Title = %q, want %q", got.Title, tt.req.Title)
				}
				if got.Status != "open" {
					t.Errorf("Status = %q, want open", got.Status)
				}
				if got.Severity != tt.req.Severity {
					t.Errorf("Severity = %q, want %q", got.Severity, tt.req.Severity)
				}
			}
		})
	}
}

func TestGetOutage(t *testing.T) {
	svc := New(newMemStorage())
	created, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "Test outage",
		Severity: "low",
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{"found", created.ID, false},
		{"not found", uuid.New(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetOutage(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetOutage() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got.ID != tt.id {
				t.Errorf("ID mismatch: got %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestListOutages(t *testing.T) {
	svc := New(newMemStorage())
	for i := 0; i < 5; i++ {
		if _, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
			Title:    "outage",
			Severity: "low",
		}); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name   string
		limit  int
		offset int
		want   int
	}{
		{"default limit", 10, 0, 5},
		{"limit 2", 2, 0, 2},
		{"offset beyond", 10, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ListOutages(context.Background(), tt.limit, tt.offset)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestUpdateOutage(t *testing.T) {
	svc := New(newMemStorage())
	created, _ := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "original",
		Severity: "low",
	})

	newTitle := "updated"
	newStatus := "investigating"
	tests := []struct {
		name    string
		id      uuid.UUID
		req     domain.UpdateOutageRequest
		wantErr bool
		check   func(*domain.Outage)
	}{
		{
			name: "update title",
			id:   created.ID,
			req:  domain.UpdateOutageRequest{Title: &newTitle},
			check: func(o *domain.Outage) {
				if o.Title != newTitle {
					t.Errorf("Title = %q, want %q", o.Title, newTitle)
				}
			},
		},
		{
			name: "update status",
			id:   created.ID,
			req:  domain.UpdateOutageRequest{Status: &newStatus},
			check: func(o *domain.Outage) {
				if o.Status != newStatus {
					t.Errorf("Status = %q, want %q", o.Status, newStatus)
				}
			},
		},
		{
			name:    "not found",
			id:      uuid.New(),
			req:     domain.UpdateOutageRequest{Title: &newTitle},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.UpdateOutage(context.Background(), tt.id, tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateOutage() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.check != nil {
				tt.check(got)
			}
		})
	}
}

func TestAddNote(t *testing.T) {
	svc := New(newMemStorage())
	created, _ := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "outage",
		Severity: "high",
	})

	tests := []struct {
		name     string
		outageID uuid.UUID
		req      domain.AddNoteRequest
		wantErr  bool
	}{
		{
			name:     "valid note",
			outageID: created.ID,
			req:      domain.AddNoteRequest{Content: "looking into it", Format: "plaintext", Author: "alice"},
		},
		{
			name:     "unknown outage",
			outageID: uuid.New(),
			req:      domain.AddNoteRequest{Content: "test", Format: "plaintext", Author: "bob"},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note, err := svc.AddNote(context.Background(), tt.outageID, tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AddNote() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if note.Content != tt.req.Content {
					t.Errorf("Content mismatch")
				}
				if note.Author != tt.req.Author {
					t.Errorf("Author mismatch")
				}
			}
		})
	}
}

func TestAddTag(t *testing.T) {
	svc := New(newMemStorage())
	created, _ := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "outage",
		Severity: "high",
	})

	tests := []struct {
		name     string
		outageID uuid.UUID
		key      string
		value    string
		wantErr  bool
	}{
		{"valid tag", created.ID, "jira", "OPS-999", false},
		{"unknown outage", uuid.New(), "jira", "OPS-000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := svc.AddTag(context.Background(), tt.outageID, tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("AddTag() err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if tag.Key != tt.key || tag.Value != tt.value {
					t.Errorf("tag = {%s:%s}, want {%s:%s}", tag.Key, tag.Value, tt.key, tt.value)
				}
			}
		})
	}
}

func TestFindOutagesByTag(t *testing.T) {
	svc := New(newMemStorage())
	o1, _ := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "A", Severity: "low"})
	o2, _ := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "B", Severity: "low"})

	if _, err := svc.AddTag(context.Background(), o1.ID, "team", "sre"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddTag(context.Background(), o2.ID, "team", "sre"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddTag(context.Background(), o2.ID, "team", "backend"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		key   string
		value string
		want  int
	}{
		{"find sre team", "team", "sre", 2},
		{"find backend team", "team", "backend", 1},
		{"no match", "team", "frontend", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.FindOutagesByTag(context.Background(), tt.key, tt.value)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tt.want {
				t.Errorf("FindOutagesByTag() len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

// Verify that memStorage satisfies the storage.Storage interface at compile time.
// We import the interface indirectly via the service package's dependency.
var _ interface {
	Close() error
} = (*memStorage)(nil)

// Verify error wrapping works.
func TestErrNotFound(t *testing.T) {
	svc := New(newMemStorage())
	_, err := svc.GetOutage(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
