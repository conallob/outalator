package grpc

import (
	"context"

	"github.com/conall/outalator/internal/service"
	pb "github.com/conallob/outalator/api/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server holds the gRPC server implementation
type Server struct {
	pb.UnimplementedOutageServiceServer
	pb.UnimplementedNoteServiceServer
	pb.UnimplementedTagServiceServer
	pb.UnimplementedAlertServiceServer
	pb.UnimplementedHealthServiceServer

	service *service.Service
}

// NewServer creates a new gRPC server
func NewServer(svc *service.Service) *Server {
	return &Server{
		service: svc,
	}
}

// RegisterServices registers all gRPC services with the gRPC server
func (s *Server) RegisterServices(grpcServer *grpc.Server) {
	pb.RegisterOutageServiceServer(grpcServer, s)
	pb.RegisterNoteServiceServer(grpcServer, s)
	pb.RegisterTagServiceServer(grpcServer, s)
	pb.RegisterAlertServiceServer(grpcServer, s)
	pb.RegisterHealthServiceServer(grpcServer, s)
}

// ============================================================================
// OutageService implementation
// ============================================================================

// CreateOutage creates a new outage with metadata and custom_fields
func (s *Server) CreateOutage(ctx context.Context, req *pb.CreateOutageRequest) (*pb.CreateOutageResponse, error) {
	// Convert request from protobuf to domain
	domainReq, err := CreateOutageRequestProtoToDomain(req)
	if err != nil {
		return nil, err
	}

	// Call service layer
	outage, err := s.service.CreateOutage(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert response from domain to protobuf
	pbOutage, err := OutageDomainToProto(outage)
	if err != nil {
		return nil, err
	}

	return &pb.CreateOutageResponse{
		Outage: pbOutage,
	}, nil
}

// GetOutage retrieves an outage by ID
func (s *Server) GetOutage(ctx context.Context, req *pb.GetOutageRequest) (*pb.GetOutageResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, err
	}

	outage, err := s.service.GetOutage(ctx, id)
	if err != nil {
		return nil, err
	}

	pbOutage, err := OutageDomainToProto(outage)
	if err != nil {
		return nil, err
	}

	return &pb.GetOutageResponse{
		Outage: pbOutage,
	}, nil
}

// ListOutages lists outages with pagination
func (s *Server) ListOutages(ctx context.Context, req *pb.ListOutagesRequest) (*pb.ListOutagesResponse, error) {
	limit := int(req.Limit)
	offset := int(req.Offset)

	outages, err := s.service.ListOutages(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	var pbOutages []*pb.Outage
	for _, outage := range outages {
		pbOutage, err := OutageDomainToProto(outage)
		if err != nil {
			return nil, err
		}
		pbOutages = append(pbOutages, pbOutage)
	}

	return &pb.ListOutagesResponse{
		Outages: pbOutages,
		Limit:   req.Limit,
		Offset:  req.Offset,
		Total:   int32(len(pbOutages)), // TODO: Get actual total count from storage
	}, nil
}

// UpdateOutage updates an outage
// NOTE: This uses FULL REPLACEMENT for metadata and custom_fields, not merging
func (s *Server) UpdateOutage(ctx context.Context, req *pb.UpdateOutageRequest) (*pb.UpdateOutageResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, err
	}

	// Convert request from protobuf to domain
	domainReq, err := UpdateOutageRequestProtoToDomain(req)
	if err != nil {
		return nil, err
	}

	// Call service layer
	outage, err := s.service.UpdateOutage(ctx, id, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert response from domain to protobuf
	pbOutage, err := OutageDomainToProto(outage)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateOutageResponse{
		Outage: pbOutage,
	}, nil
}

// DeleteOutage deletes an outage
func (s *Server) DeleteOutage(ctx context.Context, req *pb.DeleteOutageRequest) (*emptypb.Empty, error) {
	// TODO: Implement DeleteOutage in service layer
	return &emptypb.Empty{}, nil
}

// ============================================================================
// NoteService implementation
// ============================================================================

// AddNote adds a note to an outage
func (s *Server) AddNote(ctx context.Context, req *pb.AddNoteRequest) (*pb.AddNoteResponse, error) {
	outageID, err := parseUUID(req.OutageId)
	if err != nil {
		return nil, err
	}

	// Convert request from protobuf to domain
	domainReq, err := AddNoteRequestProtoToDomain(req)
	if err != nil {
		return nil, err
	}

	// Call service layer
	note, err := s.service.AddNote(ctx, outageID, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert response from domain to protobuf
	pbNote, err := NoteDomainToProto(note)
	if err != nil {
		return nil, err
	}

	return &pb.AddNoteResponse{
		Note: pbNote,
	}, nil
}

// GetNote retrieves a note by ID
func (s *Server) GetNote(ctx context.Context, req *pb.GetNoteRequest) (*pb.GetNoteResponse, error) {
	// TODO: Implement GetNote in service layer
	return &pb.GetNoteResponse{}, nil
}

// ListNotesByOutage lists notes for an outage
func (s *Server) ListNotesByOutage(ctx context.Context, req *pb.ListNotesByOutageRequest) (*pb.ListNotesByOutageResponse, error) {
	// TODO: Implement ListNotesByOutage in service layer
	return &pb.ListNotesByOutageResponse{}, nil
}

// UpdateNote updates a note
// NOTE: This uses FULL REPLACEMENT for metadata and custom_fields, not merging
func (s *Server) UpdateNote(ctx context.Context, req *pb.UpdateNoteRequest) (*pb.UpdateNoteResponse, error) {
	noteID, err := parseUUID(req.Id)
	if err != nil {
		return nil, err
	}

	// Convert request from protobuf to domain
	content, format, metadata, customFields, err := UpdateNoteRequestProtoToDomain(req)
	if err != nil {
		return nil, err
	}

	// Call service layer
	note, err := s.service.UpdateNote(ctx, noteID, content, format, metadata, customFields)
	if err != nil {
		return nil, err
	}

	// Convert response from domain to protobuf
	pbNote, err := NoteDomainToProto(note)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateNoteResponse{
		Note: pbNote,
	}, nil
}

// DeleteNote deletes a note
func (s *Server) DeleteNote(ctx context.Context, req *pb.DeleteNoteRequest) (*emptypb.Empty, error) {
	// TODO: Implement DeleteNote in service layer
	return &emptypb.Empty{}, nil
}

// ============================================================================
// TagService implementation
// ============================================================================

// AddTag adds a tag to an outage
func (s *Server) AddTag(ctx context.Context, req *pb.AddTagRequest) (*pb.AddTagResponse, error) {
	outageID, err := parseUUID(req.OutageId)
	if err != nil {
		return nil, err
	}

	customFields := protoStructToMap(req.CustomFields)

	// Call service layer
	tag, err := s.service.AddTag(ctx, outageID, req.Key, req.Value, customFields)
	if err != nil {
		return nil, err
	}

	// Convert response from domain to protobuf
	pbTag, err := TagDomainToProto(tag)
	if err != nil {
		return nil, err
	}

	return &pb.AddTagResponse{
		Tag: pbTag,
	}, nil
}

// GetTag retrieves a tag by ID
func (s *Server) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	// TODO: Implement GetTag in service layer
	return &pb.GetTagResponse{}, nil
}

// ListTagsByOutage lists tags for an outage
func (s *Server) ListTagsByOutage(ctx context.Context, req *pb.ListTagsByOutageRequest) (*pb.ListTagsByOutageResponse, error) {
	// TODO: Implement ListTagsByOutage in service layer
	return &pb.ListTagsByOutageResponse{}, nil
}

// DeleteTag deletes a tag
func (s *Server) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*emptypb.Empty, error) {
	// TODO: Implement DeleteTag in service layer
	return &emptypb.Empty{}, nil
}

// SearchOutagesByTag searches for outages by tag
func (s *Server) SearchOutagesByTag(ctx context.Context, req *pb.SearchOutagesByTagRequest) (*pb.SearchOutagesByTagResponse, error) {
	outages, err := s.service.FindOutagesByTag(ctx, req.Key, req.Value)
	if err != nil {
		return nil, err
	}

	var pbOutages []*pb.Outage
	for _, outage := range outages {
		pbOutage, err := OutageDomainToProto(outage)
		if err != nil {
			return nil, err
		}
		pbOutages = append(pbOutages, pbOutage)
	}

	return &pb.SearchOutagesByTagResponse{
		Outages: pbOutages,
	}, nil
}

// ============================================================================
// AlertService implementation
// ============================================================================

// ImportAlert imports an alert from an external service
func (s *Server) ImportAlert(ctx context.Context, req *pb.ImportAlertRequest) (*pb.ImportAlertResponse, error) {
	outageID, err := parseUUIDPtr(req.OutageId)
	if err != nil {
		return nil, err
	}

	// Call service layer
	alert, err := s.service.ImportAlert(ctx, req.Source, req.ExternalId, outageID)
	if err != nil {
		return nil, err
	}

	// Get the associated outage
	outage, err := s.service.GetOutage(ctx, alert.OutageID)
	if err != nil {
		return nil, err
	}

	// Convert to protobuf
	pbAlert, err := AlertDomainToProto(alert)
	if err != nil {
		return nil, err
	}

	pbOutage, err := OutageDomainToProto(outage)
	if err != nil {
		return nil, err
	}

	return &pb.ImportAlertResponse{
		Alert:  pbAlert,
		Outage: pbOutage,
	}, nil
}

// GetAlert retrieves an alert by ID
func (s *Server) GetAlert(ctx context.Context, req *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	// TODO: Implement GetAlert in service layer
	return &pb.GetAlertResponse{}, nil
}

// GetAlertByExternalID retrieves an alert by external ID and source
func (s *Server) GetAlertByExternalID(ctx context.Context, req *pb.GetAlertByExternalIDRequest) (*pb.GetAlertByExternalIDResponse, error) {
	// TODO: Implement GetAlertByExternalID (already exists in storage, need to expose in service)
	return &pb.GetAlertByExternalIDResponse{}, nil
}

// ListAlertsByOutage lists alerts for an outage
func (s *Server) ListAlertsByOutage(ctx context.Context, req *pb.ListAlertsByOutageRequest) (*pb.ListAlertsByOutageResponse, error) {
	// TODO: Implement ListAlertsByOutage in service layer
	return &pb.ListAlertsByOutageResponse{}, nil
}

// UpdateAlert updates an alert
// NOTE: This uses FULL REPLACEMENT for metadata and custom_fields, not merging
func (s *Server) UpdateAlert(ctx context.Context, req *pb.UpdateAlertRequest) (*pb.UpdateAlertResponse, error) {
	// TODO: Implement UpdateAlert in service layer
	return &pb.UpdateAlertResponse{}, nil
}

// ============================================================================
// HealthService implementation
// ============================================================================

// Check performs a health check
func (s *Server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: "healthy",
	}, nil
}
