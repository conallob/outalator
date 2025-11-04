package grpc

import (
	"fmt"
	"time"

	"github.com/conall/outalator/internal/domain"
	pb "github.com/conallob/outalator/api/proto/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================================================
// Helper functions for basic type conversions
// ============================================================================

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

// ============================================================================
// Helper functions for metadata and custom fields conversions
// ============================================================================

// mapToProtoStruct converts a Go map[string]any to protobuf Struct
func mapToProtoStruct(m map[string]any) (*structpb.Struct, error) {
	if m == nil {
		return nil, nil
	}
	return structpb.NewStruct(m)
}

// protoStructToMap converts a protobuf Struct to Go map[string]any
func protoStructToMap(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

// copyStringMap creates a copy of a string map (for metadata)
func copyStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ============================================================================
// Source metadata converters
// ============================================================================

// sourceMetadataToProto converts domain source_metadata map to protobuf oneof
func sourceMetadataToProto(source string, metadata map[string]any) (pb.isAlert_SourceMetadata, error) {
	if metadata == nil {
		return nil, nil
	}

	switch source {
	case "pagerduty":
		pd := &pb.PagerDutyMetadata{}
		if v, ok := metadata["incident_key"].(string); ok {
			pd.IncidentKey = v
		}
		if v, ok := metadata["service_id"].(string); ok {
			pd.ServiceId = v
		}
		if v, ok := metadata["service_name"].(string); ok {
			pd.ServiceName = v
		}
		if v, ok := metadata["escalation_policy"].(string); ok {
			pd.EscalationPolicy = v
		}
		if v, ok := metadata["assignees"].([]string); ok {
			pd.Assignees = v
		}
		if v, ok := metadata["urgency"].(string); ok {
			pd.Urgency = v
		}
		if v, ok := metadata["html_url"].(string); ok {
			pd.HtmlUrl = v
		}
		return &pb.Alert_Pagerduty{Pagerduty: pd}, nil

	case "opsgenie":
		og := &pb.OpsGenieMetadata{}
		if v, ok := metadata["alias"].(string); ok {
			og.Alias = v
		}
		if v, ok := metadata["entity"].(string); ok {
			og.Entity = v
		}
		if v, ok := metadata["responders"].([]string); ok {
			og.Responders = v
		}
		if v, ok := metadata["visible_to"].([]string); ok {
			og.VisibleTo = v
		}
		if v, ok := metadata["actions"].([]string); ok {
			og.Actions = v
		}
		if v, ok := metadata["priority"].(string); ok {
			og.Priority = v
		}
		if v, ok := metadata["owner"].(string); ok {
			og.Owner = v
		}
		return &pb.Alert_Opsgenie{Opsgenie: og}, nil

	default:
		// Generic metadata
		props := make(map[string]string)
		for k, v := range metadata {
			if str, ok := v.(string); ok {
				props[k] = str
			}
		}
		return &pb.Alert_Generic{Generic: &pb.GenericMetadata{Properties: props}}, nil
	}
}

// protoToSourceMetadata converts protobuf oneof source_metadata to domain map
func protoToSourceMetadata(sm pb.isAlert_SourceMetadata) map[string]any {
	if sm == nil {
		return nil
	}

	result := make(map[string]any)

	switch v := sm.(type) {
	case *pb.Alert_Pagerduty:
		if v.Pagerduty != nil {
			result["incident_key"] = v.Pagerduty.IncidentKey
			result["service_id"] = v.Pagerduty.ServiceId
			result["service_name"] = v.Pagerduty.ServiceName
			result["escalation_policy"] = v.Pagerduty.EscalationPolicy
			result["assignees"] = v.Pagerduty.Assignees
			result["urgency"] = v.Pagerduty.Urgency
			result["html_url"] = v.Pagerduty.HtmlUrl
		}
	case *pb.Alert_Opsgenie:
		if v.Opsgenie != nil {
			result["alias"] = v.Opsgenie.Alias
			result["entity"] = v.Opsgenie.Entity
			result["responders"] = v.Opsgenie.Responders
			result["visible_to"] = v.Opsgenie.VisibleTo
			result["actions"] = v.Opsgenie.Actions
			result["priority"] = v.Opsgenie.Priority
			result["owner"] = v.Opsgenie.Owner
		}
	case *pb.Alert_Generic:
		if v.Generic != nil {
			for k, val := range v.Generic.Properties {
				result[k] = val
			}
		}
	}

	return result
}

// ============================================================================
// Domain model converters: Outage
// ============================================================================

// OutageDomainToProto converts domain.Outage to pb.Outage
func OutageDomainToProto(o *domain.Outage) (*pb.Outage, error) {
	if o == nil {
		return nil, nil
	}

	customFields, err := mapToProtoStruct(o.CustomFields)
	if err != nil {
		return nil, fmt.Errorf("failed to convert custom_fields: %w", err)
	}

	pbOutage := &pb.Outage{
		Id:           o.ID.String(),
		Title:        o.Title,
		Description:  o.Description,
		Status:       o.Status,
		Severity:     o.Severity,
		CreatedAt:    convertTimestampToProto(o.CreatedAt),
		UpdatedAt:    convertTimestampToProto(o.UpdatedAt),
		ResolvedAt:   convertTimestampPtrToProto(o.ResolvedAt),
		Metadata:     copyStringMap(o.Metadata),
		CustomFields: customFields,
	}

	// Convert associated alerts
	for _, alert := range o.Alerts {
		pbAlert, err := AlertDomainToProto(alert)
		if err != nil {
			return nil, fmt.Errorf("failed to convert alert: %w", err)
		}
		pbOutage.Alerts = append(pbOutage.Alerts, pbAlert)
	}

	// Convert associated notes
	for _, note := range o.Notes {
		pbNote, err := NoteDomainToProto(note)
		if err != nil {
			return nil, fmt.Errorf("failed to convert note: %w", err)
		}
		pbOutage.Notes = append(pbOutage.Notes, pbNote)
	}

	// Convert associated tags
	for _, tag := range o.Tags {
		pbTag, err := TagDomainToProto(tag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tag: %w", err)
		}
		pbOutage.Tags = append(pbOutage.Tags, pbTag)
	}

	return pbOutage, nil
}

// OutageProtoToDomain converts pb.Outage to domain.Outage
func OutageProtoToDomain(pb *pb.Outage) (*domain.Outage, error) {
	if pb == nil {
		return nil, nil
	}

	id, err := parseUUID(pb.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid outage id: %w", err)
	}

	outage := &domain.Outage{
		ID:           id,
		Title:        pb.Title,
		Description:  pb.Description,
		Status:       pb.Status,
		Severity:     pb.Severity,
		CreatedAt:    convertProtoToTimestamp(pb.CreatedAt),
		UpdatedAt:    convertProtoToTimestamp(pb.UpdatedAt),
		ResolvedAt:   convertProtoToTimestampPtr(pb.ResolvedAt),
		Metadata:     copyStringMap(pb.Metadata),
		CustomFields: protoStructToMap(pb.CustomFields),
	}

	// Convert associated alerts
	for _, pbAlert := range pb.Alerts {
		alert, err := AlertProtoToDomain(pbAlert)
		if err != nil {
			return nil, fmt.Errorf("failed to convert alert: %w", err)
		}
		outage.Alerts = append(outage.Alerts, alert)
	}

	// Convert associated notes
	for _, pbNote := range pb.Notes {
		note, err := NoteProtoToDomain(pbNote)
		if err != nil {
			return nil, fmt.Errorf("failed to convert note: %w", err)
		}
		outage.Notes = append(outage.Notes, note)
	}

	// Convert associated tags
	for _, pbTag := range pb.Tags {
		tag, err := TagProtoToDomain(pbTag)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tag: %w", err)
		}
		outage.Tags = append(outage.Tags, tag)
	}

	return outage, nil
}

// ============================================================================
// Domain model converters: Alert
// ============================================================================

// AlertDomainToProto converts domain.Alert to pb.Alert
func AlertDomainToProto(a *domain.Alert) (*pb.Alert, error) {
	if a == nil {
		return nil, nil
	}

	sourceMetadata, err := sourceMetadataToProto(a.Source, a.SourceMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to convert source_metadata: %w", err)
	}

	customFields, err := mapToProtoStruct(a.CustomFields)
	if err != nil {
		return nil, fmt.Errorf("failed to convert custom_fields: %w", err)
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
		SourceMetadata:  sourceMetadata,
		Metadata:        copyStringMap(a.Metadata),
		CustomFields:    customFields,
	}, nil
}

// AlertProtoToDomain converts pb.Alert to domain.Alert
func AlertProtoToDomain(pb *pb.Alert) (*domain.Alert, error) {
	if pb == nil {
		return nil, nil
	}

	id, err := parseUUID(pb.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid alert id: %w", err)
	}

	outageID, err := parseUUID(pb.OutageId)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	return &domain.Alert{
		ID:              id,
		OutageID:        outageID,
		ExternalID:      pb.ExternalId,
		Source:          pb.Source,
		TeamName:        pb.TeamName,
		Title:           pb.Title,
		Description:     pb.Description,
		Severity:        pb.Severity,
		TriggeredAt:     convertProtoToTimestamp(pb.TriggeredAt),
		AcknowledgedAt:  convertProtoToTimestampPtr(pb.AcknowledgedAt),
		ResolvedAt:      convertProtoToTimestampPtr(pb.ResolvedAt),
		CreatedAt:       convertProtoToTimestamp(pb.CreatedAt),
		SourceMetadata:  protoToSourceMetadata(pb.SourceMetadata),
		Metadata:        copyStringMap(pb.Metadata),
		CustomFields:    protoStructToMap(pb.CustomFields),
	}, nil
}

// ============================================================================
// Domain model converters: Note
// ============================================================================

// NoteDomainToProto converts domain.Note to pb.Note
func NoteDomainToProto(n *domain.Note) (*pb.Note, error) {
	if n == nil {
		return nil, nil
	}

	customFields, err := mapToProtoStruct(n.CustomFields)
	if err != nil {
		return nil, fmt.Errorf("failed to convert custom_fields: %w", err)
	}

	return &pb.Note{
		Id:           n.ID.String(),
		OutageId:     n.OutageID.String(),
		Content:      n.Content,
		Format:       n.Format,
		Author:       n.Author,
		CreatedAt:    convertTimestampToProto(n.CreatedAt),
		UpdatedAt:    convertTimestampToProto(n.UpdatedAt),
		Metadata:     copyStringMap(n.Metadata),
		CustomFields: customFields,
	}, nil
}

// NoteProtoToDomain converts pb.Note to domain.Note
func NoteProtoToDomain(pb *pb.Note) (*domain.Note, error) {
	if pb == nil {
		return nil, nil
	}

	id, err := parseUUID(pb.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid note id: %w", err)
	}

	outageID, err := parseUUID(pb.OutageId)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	return &domain.Note{
		ID:           id,
		OutageID:     outageID,
		Content:      pb.Content,
		Format:       pb.Format,
		Author:       pb.Author,
		CreatedAt:    convertProtoToTimestamp(pb.CreatedAt),
		UpdatedAt:    convertProtoToTimestamp(pb.UpdatedAt),
		Metadata:     copyStringMap(pb.Metadata),
		CustomFields: protoStructToMap(pb.CustomFields),
	}, nil
}

// ============================================================================
// Domain model converters: Tag
// ============================================================================

// TagDomainToProto converts domain.Tag to pb.Tag
func TagDomainToProto(t *domain.Tag) (*pb.Tag, error) {
	if t == nil {
		return nil, nil
	}

	customFields, err := mapToProtoStruct(t.CustomFields)
	if err != nil {
		return nil, fmt.Errorf("failed to convert custom_fields: %w", err)
	}

	return &pb.Tag{
		Id:           t.ID.String(),
		OutageId:     t.OutageID.String(),
		Key:          t.Key,
		Value:        t.Value,
		CreatedAt:    convertTimestampToProto(t.CreatedAt),
		CustomFields: customFields,
	}, nil
}

// TagProtoToDomain converts pb.Tag to domain.Tag
func TagProtoToDomain(pb *pb.Tag) (*domain.Tag, error) {
	if pb == nil {
		return nil, nil
	}

	id, err := parseUUID(pb.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid tag id: %w", err)
	}

	outageID, err := parseUUID(pb.OutageId)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	return &domain.Tag{
		ID:           id,
		OutageID:     outageID,
		Key:          pb.Key,
		Value:        pb.Value,
		CreatedAt:    convertProtoToTimestamp(pb.CreatedAt),
		CustomFields: protoStructToMap(pb.CustomFields),
	}, nil
}

// ============================================================================
// Request/Response converters: CreateOutage
// ============================================================================

// CreateOutageRequestProtoToDomain converts pb.CreateOutageRequest to domain.CreateOutageRequest
func CreateOutageRequestProtoToDomain(pb *pb.CreateOutageRequest) (domain.CreateOutageRequest, error) {
	if pb == nil {
		return domain.CreateOutageRequest{}, fmt.Errorf("nil request")
	}

	req := domain.CreateOutageRequest{
		Title:        pb.Title,
		Description:  pb.Description,
		Severity:     pb.Severity,
		AlertIDs:     pb.AlertIds,
		Metadata:     copyStringMap(pb.Metadata),
		CustomFields: protoStructToMap(pb.CustomFields),
	}

	// Convert tags
	for _, pbTag := range pb.Tags {
		req.Tags = append(req.Tags, domain.TagInput{
			Key:          pbTag.Key,
			Value:        pbTag.Value,
			CustomFields: protoStructToMap(pbTag.CustomFields),
		})
	}

	return req, nil
}

// ============================================================================
// Request/Response converters: UpdateOutage
// ============================================================================

// UpdateOutageRequestProtoToDomain converts pb.UpdateOutageRequest to domain.UpdateOutageRequest
func UpdateOutageRequestProtoToDomain(pb *pb.UpdateOutageRequest) (domain.UpdateOutageRequest, error) {
	if pb == nil {
		return domain.UpdateOutageRequest{}, fmt.Errorf("nil request")
	}

	req := domain.UpdateOutageRequest{
		Metadata:     copyStringMap(pb.Metadata),
		CustomFields: protoStructToMap(pb.CustomFields),
	}

	if pb.Title != nil {
		req.Title = pb.Title
	}
	if pb.Description != nil {
		req.Description = pb.Description
	}
	if pb.Status != nil {
		req.Status = pb.Status
	}
	if pb.Severity != nil {
		req.Severity = pb.Severity
	}

	return req, nil
}

// ============================================================================
// Request/Response converters: AddNote
// ============================================================================

// AddNoteRequestProtoToDomain converts pb.AddNoteRequest to domain.AddNoteRequest
func AddNoteRequestProtoToDomain(pb *pb.AddNoteRequest) (domain.AddNoteRequest, error) {
	if pb == nil {
		return domain.AddNoteRequest{}, fmt.Errorf("nil request")
	}

	return domain.AddNoteRequest{
		Content:      pb.Content,
		Format:       pb.Format,
		Author:       pb.Author,
		Metadata:     copyStringMap(pb.Metadata),
		CustomFields: protoStructToMap(pb.CustomFields),
	}, nil
}

// ============================================================================
// Request/Response converters: UpdateNote
// ============================================================================

// UpdateNoteRequestProtoToDomain converts pb.UpdateNoteRequest to update parameters
func UpdateNoteRequestProtoToDomain(pb *pb.UpdateNoteRequest) (content, format *string, metadata map[string]string, customFields map[string]any, err error) {
	if pb == nil {
		return nil, nil, nil, nil, fmt.Errorf("nil request")
	}

	return pb.Content, pb.Format, copyStringMap(pb.Metadata), protoStructToMap(pb.CustomFields), nil
}
