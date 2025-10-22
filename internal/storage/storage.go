package storage

import (
	"context"

	"github.com/conall/outalator/internal/domain"
	"github.com/google/uuid"
)

// Storage defines the interface for data persistence
type Storage interface {
	OutageStorage
	AlertStorage
	NoteStorage
	TagStorage
	Close() error
}

// OutageStorage defines methods for outage persistence
type OutageStorage interface {
	CreateOutage(ctx context.Context, outage *domain.Outage) error
	GetOutage(ctx context.Context, id uuid.UUID) (*domain.Outage, error)
	ListOutages(ctx context.Context, limit, offset int) ([]*domain.Outage, error)
	UpdateOutage(ctx context.Context, outage *domain.Outage) error
	DeleteOutage(ctx context.Context, id uuid.UUID) error
}

// AlertStorage defines methods for alert persistence
type AlertStorage interface {
	CreateAlert(ctx context.Context, alert *domain.Alert) error
	GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error)
	GetAlertByExternalID(ctx context.Context, externalID, source string) (*domain.Alert, error)
	ListAlertsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Alert, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
}

// NoteStorage defines methods for note persistence
type NoteStorage interface {
	CreateNote(ctx context.Context, note *domain.Note) error
	GetNote(ctx context.Context, id uuid.UUID) (*domain.Note, error)
	ListNotesByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Note, error)
	UpdateNote(ctx context.Context, note *domain.Note) error
	DeleteNote(ctx context.Context, id uuid.UUID) error
}

// TagStorage defines methods for tag persistence
type TagStorage interface {
	CreateTag(ctx context.Context, tag *domain.Tag) error
	GetTag(ctx context.Context, id uuid.UUID) (*domain.Tag, error)
	ListTagsByOutage(ctx context.Context, outageID uuid.UUID) ([]*domain.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
	FindOutagesByTag(ctx context.Context, key, value string) ([]*domain.Outage, error)
}
