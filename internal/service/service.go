package service

import (
	"context"
	"fmt"
	"time"

	"github.com/conall/outalator/internal/domain"
	"github.com/conall/outalator/internal/notification"
	"github.com/conall/outalator/internal/storage"
	"github.com/google/uuid"
)

// Service provides business logic for the application
type Service struct {
	storage             storage.Storage
	notificationServices map[string]notification.Service
}

// New creates a new service instance
func New(storage storage.Storage) *Service {
	return &Service{
		storage:             storage,
		notificationServices: make(map[string]notification.Service),
	}
}

// RegisterNotificationService registers a notification service
func (s *Service) RegisterNotificationService(svc notification.Service) {
	s.notificationServices[svc.Name()] = svc
}

// CreateOutage creates a new outage with associated alerts
func (s *Service) CreateOutage(ctx context.Context, req domain.CreateOutageRequest) (*domain.Outage, error) {
	now := time.Now()
	outageID := uuid.New()

	outage := &domain.Outage{
		ID:          outageID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "open",
		Severity:    req.Severity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.storage.CreateOutage(ctx, outage); err != nil {
		return nil, fmt.Errorf("failed to create outage: %w", err)
	}

	// Import alerts if provided
	for _, alertID := range req.AlertIDs {
		// Try to fetch from each notification service
		for _, svc := range s.notificationServices {
			notifAlert, err := svc.FetchAlert(ctx, alertID)
			if err != nil {
				continue // Try next service
			}

			alert := &domain.Alert{
				ID:             uuid.New(),
				OutageID:       outageID,
				ExternalID:     notifAlert.ExternalID,
				Source:         notifAlert.Source,
				TeamName:       notifAlert.TeamName,
				Title:          notifAlert.Title,
				Description:    notifAlert.Description,
				Severity:       notifAlert.Severity,
				TriggeredAt:    notifAlert.TriggeredAt,
				AcknowledgedAt: notifAlert.AcknowledgedAt,
				ResolvedAt:     notifAlert.ResolvedAt,
				CreatedAt:      now,
			}

			if err := s.storage.CreateAlert(ctx, alert); err != nil {
				return nil, fmt.Errorf("failed to create alert: %w", err)
			}
			break
		}
	}

	// Create tags
	for _, tagReq := range req.Tags {
		tag := &domain.Tag{
			ID:        uuid.New(),
			OutageID:  outageID,
			Key:       tagReq.Key,
			Value:     tagReq.Value,
			CreatedAt: now,
		}
		if err := s.storage.CreateTag(ctx, tag); err != nil {
			return nil, fmt.Errorf("failed to create tag: %w", err)
		}
	}

	// Reload outage with all associations
	return s.storage.GetOutage(ctx, outageID)
}

// GetOutage retrieves an outage by ID
func (s *Service) GetOutage(ctx context.Context, id uuid.UUID) (*domain.Outage, error) {
	return s.storage.GetOutage(ctx, id)
}

// ListOutages retrieves a paginated list of outages
func (s *Service) ListOutages(ctx context.Context, limit, offset int) ([]*domain.Outage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.storage.ListOutages(ctx, limit, offset)
}

// UpdateOutage updates an outage
func (s *Service) UpdateOutage(ctx context.Context, id uuid.UUID, req domain.UpdateOutageRequest) (*domain.Outage, error) {
	outage, err := s.storage.GetOutage(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		outage.Title = *req.Title
	}
	if req.Description != nil {
		outage.Description = *req.Description
	}
	if req.Status != nil {
		outage.Status = *req.Status
		if *req.Status == "resolved" || *req.Status == "closed" {
			now := time.Now()
			outage.ResolvedAt = &now
		}
	}
	if req.Severity != nil {
		outage.Severity = *req.Severity
	}

	outage.UpdatedAt = time.Now()

	if err := s.storage.UpdateOutage(ctx, outage); err != nil {
		return nil, err
	}

	return s.storage.GetOutage(ctx, id)
}

// AddNote adds a note to an outage
func (s *Service) AddNote(ctx context.Context, outageID uuid.UUID, req domain.AddNoteRequest) (*domain.Note, error) {
	// Verify outage exists
	if _, err := s.storage.GetOutage(ctx, outageID); err != nil {
		return nil, err
	}

	now := time.Now()
	note := &domain.Note{
		ID:        uuid.New(),
		OutageID:  outageID,
		Content:   req.Content,
		Format:    req.Format,
		Author:    req.Author,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.storage.CreateNote(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

// AddTag adds a tag to an outage
func (s *Service) AddTag(ctx context.Context, outageID uuid.UUID, key, value string) (*domain.Tag, error) {
	// Verify outage exists
	if _, err := s.storage.GetOutage(ctx, outageID); err != nil {
		return nil, err
	}

	tag := &domain.Tag{
		ID:        uuid.New(),
		OutageID:  outageID,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
	}

	if err := s.storage.CreateTag(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

// FindOutagesByTag finds outages with a specific tag
func (s *Service) FindOutagesByTag(ctx context.Context, key, value string) ([]*domain.Outage, error) {
	return s.storage.FindOutagesByTag(ctx, key, value)
}

// ImportAlert imports an alert from a notification service
func (s *Service) ImportAlert(ctx context.Context, source, externalID string, outageID *uuid.UUID) (*domain.Alert, error) {
	svc, ok := s.notificationServices[source]
	if !ok {
		return nil, fmt.Errorf("notification service %s not found", source)
	}

	notifAlert, err := svc.FetchAlert(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alert from %s: %w", source, err)
	}

	// Check if alert already exists
	existing, err := s.storage.GetAlertByExternalID(ctx, externalID, source)
	if err == nil {
		return existing, nil
	}

	// Determine outage ID
	var finalOutageID uuid.UUID
	if outageID != nil {
		finalOutageID = *outageID
	} else {
		// Create a new outage for this alert
		outage := &domain.Outage{
			ID:          uuid.New(),
			Title:       notifAlert.Title,
			Description: notifAlert.Description,
			Status:      "open",
			Severity:    notifAlert.Severity,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.storage.CreateOutage(ctx, outage); err != nil {
			return nil, fmt.Errorf("failed to create outage: %w", err)
		}
		finalOutageID = outage.ID
	}

	alert := &domain.Alert{
		ID:             uuid.New(),
		OutageID:       finalOutageID,
		ExternalID:     notifAlert.ExternalID,
		Source:         notifAlert.Source,
		TeamName:       notifAlert.TeamName,
		Title:          notifAlert.Title,
		Description:    notifAlert.Description,
		Severity:       notifAlert.Severity,
		TriggeredAt:    notifAlert.TriggeredAt,
		AcknowledgedAt: notifAlert.AcknowledgedAt,
		ResolvedAt:     notifAlert.ResolvedAt,
		CreatedAt:      time.Now(),
	}

	if err := s.storage.CreateAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return alert, nil
}
