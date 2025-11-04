package grpc

import (
	"github.com/conall/outalator/internal/service"
	"google.golang.org/grpc"
)

// Server holds the gRPC server implementation
type Server struct {
	// UnimplementedOutageServiceServer - will be added after proto generation
	// UnimplementedNoteServiceServer
	// UnimplementedTagServiceServer
	// UnimplementedAlertServiceServer
	// UnimplementedHealthServiceServer

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
	// After proto generation, register services here:
	// pb.RegisterOutageServiceServer(grpcServer, s)
	// pb.RegisterNoteServiceServer(grpcServer, s)
	// pb.RegisterTagServiceServer(grpcServer, s)
	// pb.RegisterAlertServiceServer(grpcServer, s)
	// pb.RegisterHealthServiceServer(grpcServer, s)
}

// Example implementation structure (to be completed after proto generation):

/*
// CreateOutage implements OutageService.CreateOutage
func (s *Server) CreateOutage(ctx context.Context, req *pb.CreateOutageRequest) (*pb.CreateOutageResponse, error) {
	// Convert proto request to domain request
	domainReq := CreateOutageRequestProtoToDomain(req)

	// Call service layer
	outage, err := s.service.CreateOutage(ctx, domainReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create outage: %v", err)
	}

	// Convert domain response to proto response
	return &pb.CreateOutageResponse{
		Outage: OutageDomainToProto(outage),
	}, nil
}

// GetOutage implements OutageService.GetOutage
func (s *Server) GetOutage(ctx context.Context, req *pb.GetOutageRequest) (*pb.GetOutageResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
	}

	outage, err := s.service.GetOutage(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "outage not found: %v", err)
	}

	return &pb.GetOutageResponse{
		Outage: OutageDomainToProto(outage),
	}, nil
}

// ListOutages implements OutageService.ListOutages
func (s *Server) ListOutages(ctx context.Context, req *pb.ListOutagesRequest) (*pb.ListOutagesResponse, error) {
	limit := int(req.Limit)
	offset := int(req.Offset)

	outages, err := s.service.ListOutages(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list outages: %v", err)
	}

	var pbOutages []*pb.Outage
	for _, outage := range outages {
		pbOutages = append(pbOutages, OutageDomainToProto(outage))
	}

	return &pb.ListOutagesResponse{
		Outages: pbOutages,
		Limit:   req.Limit,
		Offset:  req.Offset,
		Total:   int32(len(outages)), // TODO: Get actual total count
	}, nil
}

// UpdateOutage implements OutageService.UpdateOutage
func (s *Server) UpdateOutage(ctx context.Context, req *pb.UpdateOutageRequest) (*pb.UpdateOutageResponse, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
	}

	domainReq := domain.UpdateOutageRequest{
		Title:       stringPtr(req.GetTitle()),
		Description: stringPtr(req.GetDescription()),
		Status:      stringPtr(req.GetStatus()),
		Severity:    stringPtr(req.GetSeverity()),
	}

	outage, err := s.service.UpdateOutage(ctx, id, domainReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update outage: %v", err)
	}

	return &pb.UpdateOutageResponse{
		Outage: OutageDomainToProto(outage),
	}, nil
}

// DeleteOutage implements OutageService.DeleteOutage
func (s *Server) DeleteOutage(ctx context.Context, req *pb.DeleteOutageRequest) (*emptypb.Empty, error) {
	id, err := parseUUID(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
	}

	// TODO: Implement DeleteOutage in service layer
	_ = id

	return &emptypb.Empty{}, status.Errorf(codes.Unimplemented, "DeleteOutage not yet implemented")
}

// AddNote implements NoteService.AddNote
func (s *Server) AddNote(ctx context.Context, req *pb.AddNoteRequest) (*pb.AddNoteResponse, error) {
	outageID, err := parseUUID(req.OutageId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
	}

	domainReq := domain.AddNoteRequest{
		Content: req.Content,
		Format:  req.Format,
		Author:  req.Author,
	}

	note, err := s.service.AddNote(ctx, outageID, domainReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add note: %v", err)
	}

	return &pb.AddNoteResponse{
		Note: NoteDomainToProto(note),
	}, nil
}

// AddTag implements TagService.AddTag
func (s *Server) AddTag(ctx context.Context, req *pb.AddTagRequest) (*pb.AddTagResponse, error) {
	outageID, err := parseUUID(req.OutageId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
	}

	tag, err := s.service.AddTag(ctx, outageID, req.Key, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add tag: %v", err)
	}

	return &pb.AddTagResponse{
		Tag: TagDomainToProto(tag),
	}, nil
}

// SearchOutagesByTag implements TagService.SearchOutagesByTag
func (s *Server) SearchOutagesByTag(ctx context.Context, req *pb.SearchOutagesByTagRequest) (*pb.SearchOutagesByTagResponse, error) {
	outages, err := s.service.FindOutagesByTag(ctx, req.Key, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search outages: %v", err)
	}

	var pbOutages []*pb.Outage
	for _, outage := range outages {
		pbOutages = append(pbOutages, OutageDomainToProto(outage))
	}

	return &pb.SearchOutagesByTagResponse{
		Outages: pbOutages,
	}, nil
}

// ImportAlert implements AlertService.ImportAlert
func (s *Server) ImportAlert(ctx context.Context, req *pb.ImportAlertRequest) (*pb.ImportAlertResponse, error) {
	var outageID *uuid.UUID
	var err error

	if req.OutageId != nil {
		outageID, err = parseUUIDPtr(*req.OutageId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid outage ID: %v", err)
		}
	}

	alert, err := s.service.ImportAlert(ctx, req.Source, req.ExternalId, outageID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to import alert: %v", err)
	}

	// Get the associated outage
	outage, err := s.service.GetOutage(ctx, alert.OutageID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get outage: %v", err)
	}

	return &pb.ImportAlertResponse{
		Alert:  AlertDomainToProto(alert),
		Outage: OutageDomainToProto(outage),
	}, nil
}

// Check implements HealthService.Check
func (s *Server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: "healthy",
	}, nil
}
*/
