//go:build sqlite

package sqlite_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/conall/outalator/internal/domain"
	"github.com/conall/outalator/internal/storage/sqlite"
	"github.com/google/uuid"
)

// newStore opens a fresh in-memory database for each test.
func newStore(t *testing.T) *sqlite.SQLiteStorage {
	t.Helper()
	store, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("sqlite.New: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func now() time.Time { return time.Now().UTC().Truncate(time.Second) }

// ── Outage ────────────────────────────────────────────────────────────────────

func TestOutage_CRUD(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID:          uuid.New(),
		Title:       "Database latency spike",
		Description: "P99 latency > 5 s on primary",
		Status:      "open",
		Severity:    "high",
		CreatedAt:   now(),
		UpdatedAt:   now(),
		Metadata:    map[string]string{"team": "platform"},
	}

	// Create
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	// Get — full record with eager-loaded slices present (empty, not nil)
	got, err := s.GetOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("GetOutage: %v", err)
	}
	if got.Title != outage.Title {
		t.Errorf("Title: got %q, want %q", got.Title, outage.Title)
	}
	if got.Metadata["team"] != "platform" {
		t.Errorf("Metadata[team]: got %q, want %q", got.Metadata["team"], "platform")
	}

	// Update
	outage.Title = "Database latency spike — resolved"
	outage.Status = "resolved"
	outage.UpdatedAt = now()
	if err := s.UpdateOutage(ctx, outage); err != nil {
		t.Fatalf("UpdateOutage: %v", err)
	}
	got, err = s.GetOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("GetOutage after update: %v", err)
	}
	if got.Status != "resolved" {
		t.Errorf("Status: got %q, want %q", got.Status, "resolved")
	}

	// List
	list, err := s.ListOutages(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListOutages: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListOutages count: got %d, want 1", len(list))
	}

	// Delete
	if err := s.DeleteOutage(ctx, outage.ID); err != nil {
		t.Fatalf("DeleteOutage: %v", err)
	}

	// Not-found after delete
	_, err = s.GetOutage(ctx, outage.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetOutage after delete: got %v, want domain.ErrNotFound", err)
	}
}

func TestOutage_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	_, err := s.GetOutage(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetOutage missing: got %v, want domain.ErrNotFound", err)
	}

	err = s.UpdateOutage(ctx, &domain.Outage{ID: uuid.New(), UpdatedAt: now()})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("UpdateOutage missing: got %v, want domain.ErrNotFound", err)
	}

	err = s.DeleteOutage(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("DeleteOutage missing: got %v, want domain.ErrNotFound", err)
	}
}

func TestOutage_ListPagination(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	for i := 0; i < 5; i++ {
		o := &domain.Outage{
			ID:        uuid.New(),
			Title:     "outage",
			Status:    "open",
			Severity:  "low",
			CreatedAt: now(),
			UpdatedAt: now(),
		}
		if err := s.CreateOutage(ctx, o); err != nil {
			t.Fatalf("CreateOutage %d: %v", i, err)
		}
	}

	page1, err := s.ListOutages(ctx, 3, 0)
	if err != nil {
		t.Fatalf("ListOutages page 1: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("page 1 count: got %d, want 3", len(page1))
	}

	page2, err := s.ListOutages(ctx, 3, 3)
	if err != nil {
		t.Fatalf("ListOutages page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("page 2 count: got %d, want 2", len(page2))
	}
}

// ── Alert ─────────────────────────────────────────────────────────────────────

func TestAlert_CRUD(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID: uuid.New(), Title: "o", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	alert := &domain.Alert{
		ID:          uuid.New(),
		OutageID:    outage.ID,
		ExternalID:  "pd-12345",
		Source:      "pagerduty",
		Title:       "Latency alert",
		Severity:    "high",
		TriggeredAt: now(),
		CreatedAt:   now(),
		Metadata:    map[string]string{"service": "api"},
	}

	// Create
	if err := s.CreateAlert(ctx, alert); err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}

	// Get by ID
	got, err := s.GetAlert(ctx, alert.ID)
	if err != nil {
		t.Fatalf("GetAlert: %v", err)
	}
	if got.ExternalID != alert.ExternalID {
		t.Errorf("ExternalID: got %q, want %q", got.ExternalID, alert.ExternalID)
	}
	if got.Metadata["service"] != "api" {
		t.Errorf("Metadata[service]: got %q, want %q", got.Metadata["service"], "api")
	}

	// Get by external ID
	got, err = s.GetAlertByExternalID(ctx, alert.ExternalID, alert.Source)
	if err != nil {
		t.Fatalf("GetAlertByExternalID: %v", err)
	}
	if got.ID != alert.ID {
		t.Errorf("ID mismatch: got %s, want %s", got.ID, alert.ID)
	}

	// List by outage
	list, err := s.ListAlertsByOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("ListAlertsByOutage: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list count: got %d, want 1", len(list))
	}

	// Update
	alert.Title = "Latency alert — ack"
	if err := s.UpdateAlert(ctx, alert); err != nil {
		t.Fatalf("UpdateAlert: %v", err)
	}
	got, err = s.GetAlert(ctx, alert.ID)
	if err != nil {
		t.Fatalf("GetAlert after update: %v", err)
	}
	if got.Title != "Latency alert — ack" {
		t.Errorf("Title after update: got %q, want %q", got.Title, "Latency alert — ack")
	}
}

func TestAlert_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	_, err := s.GetAlert(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetAlert missing: got %v, want domain.ErrNotFound", err)
	}

	_, err = s.GetAlertByExternalID(ctx, "nope", "pagerduty")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetAlertByExternalID missing: got %v, want domain.ErrNotFound", err)
	}
}

func TestAlert_CascadeDeleteWithOutage(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID: uuid.New(), Title: "o", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	alert := &domain.Alert{
		ID: uuid.New(), OutageID: outage.ID,
		ExternalID: "x", Source: "opsgenie",
		Title: "t", TriggeredAt: now(), CreatedAt: now(),
	}
	if err := s.CreateAlert(ctx, alert); err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}

	// Deleting the outage should cascade-delete the alert.
	if err := s.DeleteOutage(ctx, outage.ID); err != nil {
		t.Fatalf("DeleteOutage: %v", err)
	}
	_, err := s.GetAlert(ctx, alert.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetAlert after cascade delete: got %v, want domain.ErrNotFound", err)
	}
}

// ── Note ──────────────────────────────────────────────────────────────────────

func TestNote_CRUD(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID: uuid.New(), Title: "o", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	note := &domain.Note{
		ID:        uuid.New(),
		OutageID:  outage.ID,
		Content:   "Checked dashboards — latency from CDN edge.",
		Format:    "plaintext",
		Author:    "alice",
		CreatedAt: now(),
		UpdatedAt: now(),
	}

	// Create
	if err := s.CreateNote(ctx, note); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}

	// Get
	got, err := s.GetNote(ctx, note.ID)
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if got.Content != note.Content {
		t.Errorf("Content: got %q, want %q", got.Content, note.Content)
	}
	if got.Author != "alice" {
		t.Errorf("Author: got %q, want %q", got.Author, "alice")
	}

	// List
	list, err := s.ListNotesByOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("ListNotesByOutage: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list count: got %d, want 1", len(list))
	}

	// Update — content changes, author must stay unchanged
	note.Content = "Root cause: misconfigured CDN TTL."
	note.UpdatedAt = now()
	if err := s.UpdateNote(ctx, note); err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	got, err = s.GetNote(ctx, note.ID)
	if err != nil {
		t.Fatalf("GetNote after update: %v", err)
	}
	if got.Content != "Root cause: misconfigured CDN TTL." {
		t.Errorf("Content after update: got %q", got.Content)
	}
	if got.Author != "alice" {
		t.Errorf("Author changed unexpectedly: got %q", got.Author)
	}

	// Delete
	if err := s.DeleteNote(ctx, note.ID); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	_, err = s.GetNote(ctx, note.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetNote after delete: got %v, want domain.ErrNotFound", err)
	}
}

func TestNote_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	_, err := s.GetNote(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetNote missing: got %v, want domain.ErrNotFound", err)
	}

	err = s.DeleteNote(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("DeleteNote missing: got %v, want domain.ErrNotFound", err)
	}
}

// ── Tag ───────────────────────────────────────────────────────────────────────

func TestTag_CRUD(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID: uuid.New(), Title: "o", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	tag := &domain.Tag{
		ID:        uuid.New(),
		OutageID:  outage.ID,
		Key:       "env",
		Value:     "production",
		CreatedAt: now(),
	}

	// Create
	if err := s.CreateTag(ctx, tag); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	// Get
	got, err := s.GetTag(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetTag: %v", err)
	}
	if got.Key != "env" || got.Value != "production" {
		t.Errorf("Key/Value: got %q=%q, want env=production", got.Key, got.Value)
	}

	// List
	list, err := s.ListTagsByOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("ListTagsByOutage: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list count: got %d, want 1", len(list))
	}

	// Delete
	if err := s.DeleteTag(ctx, tag.ID); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	_, err = s.GetTag(ctx, tag.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetTag after delete: got %v, want domain.ErrNotFound", err)
	}
}

func TestTag_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	_, err := s.GetTag(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetTag missing: got %v, want domain.ErrNotFound", err)
	}

	err = s.DeleteTag(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("DeleteTag missing: got %v, want domain.ErrNotFound", err)
	}
}

func TestFindOutagesByTag(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	// Create two outages; tag only one.
	o1 := &domain.Outage{
		ID: uuid.New(), Title: "o1", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	o2 := &domain.Outage{
		ID: uuid.New(), Title: "o2", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	for _, o := range []*domain.Outage{o1, o2} {
		if err := s.CreateOutage(ctx, o); err != nil {
			t.Fatalf("CreateOutage: %v", err)
		}
	}

	tag := &domain.Tag{
		ID:        uuid.New(),
		OutageID:  o1.ID,
		Key:       "region",
		Value:     "us-east-1",
		CreatedAt: now(),
	}
	if err := s.CreateTag(ctx, tag); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	results, err := s.FindOutagesByTag(ctx, "region", "us-east-1")
	if err != nil {
		t.Fatalf("FindOutagesByTag: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("FindOutagesByTag count: got %d, want 1", len(results))
	}
	if results[0].ID != o1.ID {
		t.Errorf("FindOutagesByTag: got outage %s, want %s", results[0].ID, o1.ID)
	}

	// Non-matching tag returns empty, not error.
	results, err = s.FindOutagesByTag(ctx, "region", "eu-west-1")
	if err != nil {
		t.Fatalf("FindOutagesByTag (no match): %v", err)
	}
	if len(results) != 0 {
		t.Errorf("FindOutagesByTag (no match) count: got %d, want 0", len(results))
	}
}

// ── GetOutage eager-loading ───────────────────────────────────────────────────

func TestGetOutage_EagerLoadsRelations(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID: uuid.New(), Title: "o", Status: "open", Severity: "low",
		CreatedAt: now(), UpdatedAt: now(),
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage: %v", err)
	}

	alert := &domain.Alert{
		ID: uuid.New(), OutageID: outage.ID,
		ExternalID: "x", Source: "pagerduty",
		Title: "a", TriggeredAt: now(), CreatedAt: now(),
	}
	note := &domain.Note{
		ID: uuid.New(), OutageID: outage.ID,
		Content: "n", Format: "plaintext", Author: "bob",
		CreatedAt: now(), UpdatedAt: now(),
	}
	tag := &domain.Tag{
		ID: uuid.New(), OutageID: outage.ID,
		Key: "k", Value: "v", CreatedAt: now(),
	}

	if err := s.CreateAlert(ctx, alert); err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}
	if err := s.CreateNote(ctx, note); err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if err := s.CreateTag(ctx, tag); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}

	got, err := s.GetOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("GetOutage: %v", err)
	}
	if len(got.Alerts) != 1 {
		t.Errorf("Alerts: got %d, want 1", len(got.Alerts))
	}
	if len(got.Notes) != 1 {
		t.Errorf("Notes: got %d, want 1", len(got.Notes))
	}
	if len(got.Tags) != 1 {
		t.Errorf("Tags: got %d, want 1", len(got.Tags))
	}
}

// ── marshalJSONAny nil handling ───────────────────────────────────────────────

func TestOutage_NilMetadata(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	outage := &domain.Outage{
		ID:        uuid.New(),
		Title:     "nil-meta",
		Status:    "open",
		Severity:  "low",
		CreatedAt: now(),
		UpdatedAt: now(),
		Metadata:  nil, // explicitly nil
	}
	if err := s.CreateOutage(ctx, outage); err != nil {
		t.Fatalf("CreateOutage with nil metadata: %v", err)
	}
	got, err := s.GetOutage(ctx, outage.ID)
	if err != nil {
		t.Fatalf("GetOutage: %v", err)
	}
	// Nil metadata should round-trip as an empty (or nil) map, not cause an error.
	_ = got
}
