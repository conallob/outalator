// Package testutil provides shared test helpers for outalator packages.
package testutil

import (
	"context"
	"sort"
	"sync"

	"github.com/conall/outalator/domain"
	"github.com/conall/outalator/storage"
	"github.com/google/uuid"
)

// Compile-time assertion that MemStorage satisfies the full storage.Storage interface.
var _ storage.Storage = (*MemStorage)(nil)

// MemStorage is a thread-safe in-memory implementation of storage.Storage for tests.
// All collections are sorted deterministically (by ID) so that pagination is stable.
type MemStorage struct {
	mu      sync.RWMutex
	outages map[uuid.UUID]*domain.Outage
	notes   map[uuid.UUID]*domain.Note
	tags    map[uuid.UUID]*domain.Tag
	alerts  map[uuid.UUID]*domain.Alert
}

// NewMemStorage returns an empty MemStorage ready for use in tests.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		outages: make(map[uuid.UUID]*domain.Outage),
		notes:   make(map[uuid.UUID]*domain.Note),
		tags:    make(map[uuid.UUID]*domain.Tag),
		alerts:  make(map[uuid.UUID]*domain.Alert),
	}
}

func (m *MemStorage) Close() error { return nil }

// --- Outage ---

func (m *MemStorage) CreateOutage(_ context.Context, o *domain.Outage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *o
	m.outages[o.ID] = &cp
	return nil
}

func (m *MemStorage) GetOutage(_ context.Context, id uuid.UUID) (*domain.Outage, error) {
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

// ListOutages returns outages sorted by ID for deterministic pagination.
func (m *MemStorage) ListOutages(_ context.Context, limit, offset int) ([]*domain.Outage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]*domain.Outage, 0, len(m.outages))
	for _, o := range m.outages {
		cp := *o
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].ID.String() < all[j].ID.String()
	})
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MemStorage) UpdateOutage(_ context.Context, o *domain.Outage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.outages[o.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *o
	m.outages[o.ID] = &cp
	return nil
}

// DeleteOutage removes the outage but does NOT cascade to notes, tags, or alerts.
// This diverges from the Postgres implementation (which cascades via FK constraints).
// Tests that delete an outage and then query its associated entities will see stale data.
func (m *MemStorage) DeleteOutage(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.outages, id)
	return nil
}

// --- Alert ---

func (m *MemStorage) CreateAlert(_ context.Context, a *domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *a
	m.alerts[a.ID] = &cp
	return nil
}

func (m *MemStorage) GetAlert(_ context.Context, id uuid.UUID) (*domain.Alert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.alerts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *a
	return &cp, nil
}

func (m *MemStorage) GetAlertByExternalID(_ context.Context, externalID, source string) (*domain.Alert, error) {
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

func (m *MemStorage) ListAlertsByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Alert, error) {
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

func (m *MemStorage) UpdateAlert(_ context.Context, a *domain.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.alerts[a.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *a
	m.alerts[a.ID] = &cp
	return nil
}

// --- Note ---

func (m *MemStorage) CreateNote(_ context.Context, n *domain.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *n
	m.notes[n.ID] = &cp
	return nil
}

func (m *MemStorage) GetNote(_ context.Context, id uuid.UUID) (*domain.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.notes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *n
	return &cp, nil
}

func (m *MemStorage) ListNotesByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Note, error) {
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

func (m *MemStorage) UpdateNote(_ context.Context, n *domain.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[n.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *n
	m.notes[n.ID] = &cp
	return nil
}

func (m *MemStorage) DeleteNote(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.notes, id)
	return nil
}

// --- Tag ---

func (m *MemStorage) CreateTag(_ context.Context, t *domain.Tag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *t
	m.tags[t.ID] = &cp
	return nil
}

func (m *MemStorage) GetTag(_ context.Context, id uuid.UUID) (*domain.Tag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tags[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *MemStorage) ListTagsByOutage(_ context.Context, outageID uuid.UUID) ([]*domain.Tag, error) {
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

func (m *MemStorage) DeleteTag(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tags, id)
	return nil
}

func (m *MemStorage) FindOutagesByTag(_ context.Context, key, value string) ([]*domain.Outage, error) {
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
