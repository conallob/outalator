package service

import (
	"context"
	"errors"
	"testing"

	"github.com/conall/outalator/domain"
	"github.com/conall/outalator/internal/testutil"
	"github.com/conall/outalator/storage"
	"github.com/google/uuid"
)

// Compile-time assertion: testutil.MemStorage satisfies storage.Storage.
var _ storage.Storage = (*testutil.MemStorage)(nil)

func newSvc() *Service {
	return New(testutil.NewMemStorage())
}

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
			svc := newSvc()
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
	svc := newSvc()
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
	svc := newSvc()
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
	newTitle := "updated"
	newStatus := "investigating"
	tests := []struct {
		name    string
		req     domain.UpdateOutageRequest
		wantErr bool
		check   func(*domain.Outage)
	}{
		{
			name: "update title",
			req:  domain.UpdateOutageRequest{Title: &newTitle},
			check: func(o *domain.Outage) {
				if o.Title != newTitle {
					t.Errorf("Title = %q, want %q", o.Title, newTitle)
				}
			},
		},
		{
			name: "update status",
			req:  domain.UpdateOutageRequest{Status: &newStatus},
			check: func(o *domain.Outage) {
				if o.Status != newStatus {
					t.Errorf("Status = %q, want %q", o.Status, newStatus)
				}
			},
		},
		{
			name:    "not found",
			req:     domain.UpdateOutageRequest{Title: &newTitle},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc()
			// Create a fresh outage for each sub-test to avoid order dependence.
			var id uuid.UUID
			if !tt.wantErr {
				o, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
					Title:    "original",
					Severity: "low",
				})
				if err != nil {
					t.Fatal(err)
				}
				id = o.ID
			} else {
				id = uuid.New() // non-existent
			}

			got, err := svc.UpdateOutage(context.Background(), id, tt.req)
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
	svc := newSvc()
	created, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "outage",
		Severity: "high",
	})
	if err != nil {
		t.Fatal(err)
	}

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
	svc := newSvc()
	created, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{
		Title:    "outage",
		Severity: "high",
	})
	if err != nil {
		t.Fatal(err)
	}

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
	svc := newSvc()
	o1, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "A", Severity: "low"})
	if err != nil {
		t.Fatal(err)
	}
	o2, err := svc.CreateOutage(context.Background(), domain.CreateOutageRequest{Title: "B", Severity: "low"})
	if err != nil {
		t.Fatal(err)
	}

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

func TestErrNotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.GetOutage(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
