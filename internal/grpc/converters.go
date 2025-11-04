package grpc

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note: This file will contain conversions between domain models and protobuf messages
// After running protoc, import the generated pb package here

// Helper functions for converting between domain models and protobuf messages
// These will be implemented once the proto files are generated

// convertTimestampToProto converts a time.Time to protobuf timestamp
func convertTimestampToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

// convertTimestampPtrToProto converts a *time.Time to protobuf timestamp
func convertTimestampPtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// convertProtoToTimestamp converts a protobuf timestamp to time.Time
func convertProtoToTimestamp(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

// convertProtoToTimestampPtr converts a protobuf timestamp to *time.Time
func convertProtoToTimestampPtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	ts := t.AsTime()
	return &ts
}

// parseUUID parses a string UUID and returns error if invalid
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// parseUUIDPtr parses a string UUID pointer
func parseUUIDPtr(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Example conversion functions (to be completed after proto generation):

/*
// OutageDomainToProto converts a domain.Outage to pb.Outage
func OutageDomainToProto(o *domain.Outage) *pb.Outage {
	if o == nil {
		return nil
	}

	pbOutage := &pb.Outage{
		Id:          o.ID.String(),
		Title:       o.Title,
		Description: o.Description,
		Status:      o.Status,
		Severity:    o.Severity,
		CreatedAt:   convertTimestampToProto(o.CreatedAt),
		UpdatedAt:   convertTimestampToProto(o.UpdatedAt),
		ResolvedAt:  convertTimestampPtrToProto(o.ResolvedAt),
	}

	// Convert alerts
	for _, alert := range o.Alerts {
		pbOutage.Alerts = append(pbOutage.Alerts, AlertDomainToProto(&alert))
	}

	// Convert notes
	for _, note := range o.Notes {
		pbOutage.Notes = append(pbOutage.Notes, NoteDomainToProto(&note))
	}

	// Convert tags
	for _, tag := range o.Tags {
		pbOutage.Tags = append(pbOutage.Tags, TagDomainToProto(&tag))
	}

	return pbOutage
}

// AlertDomainToProto converts a domain.Alert to pb.Alert
func AlertDomainToProto(a *domain.Alert) *pb.Alert {
	if a == nil {
		return nil
	}

	return &pb.Alert{
		Id:              a.ID.String(),
		OutageId:        a.OutageID.String(),
		ExternalId:      a.ExternalID,
		Source:          a.Source,
		TeamName:        a.TeamName,
		Title:           a.Title,
		Description:     a.Description,
		Severity:        a.Severity,
		TriggeredAt:     convertTimestampToProto(a.TriggeredAt),
		AcknowledgedAt:  convertTimestampPtrToProto(a.AcknowledgedAt),
		ResolvedAt:      convertTimestampPtrToProto(a.ResolvedAt),
		CreatedAt:       convertTimestampToProto(a.CreatedAt),
	}
}

// NoteDomainToProto converts a domain.Note to pb.Note
func NoteDomainToProto(n *domain.Note) *pb.Note {
	if n == nil {
		return nil
	}

	return &pb.Note{
		Id:        n.ID.String(),
		OutageId:  n.OutageID.String(),
		Content:   n.Content,
		Format:    n.Format,
		Author:    n.Author,
		CreatedAt: convertTimestampToProto(n.CreatedAt),
		UpdatedAt: convertTimestampToProto(n.UpdatedAt),
	}
}

// TagDomainToProto converts a domain.Tag to pb.Tag
func TagDomainToProto(t *domain.Tag) *pb.Tag {
	if t == nil {
		return nil
	}

	return &pb.Tag{
		Id:        t.ID.String(),
		OutageId:  t.OutageID.String(),
		Key:       t.Key,
		Value:     t.Value,
		CreatedAt: convertTimestampToProto(t.CreatedAt),
	}
}

// CreateOutageRequestProtoToDomain converts pb.CreateOutageRequest to domain.CreateOutageRequest
func CreateOutageRequestProtoToDomain(req *pb.CreateOutageRequest) domain.CreateOutageRequest {
	domainReq := domain.CreateOutageRequest{
		Title:       req.Title,
		Description: req.Description,
		Severity:    req.Severity,
		AlertIDs:    req.AlertIds,
	}

	for _, tag := range req.Tags {
		domainReq.Tags = append(domainReq.Tags, domain.TagInput{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}

	return domainReq
}
*/
