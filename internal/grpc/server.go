package grpc

import (
	"github.com/conall/outalator/internal/service"
	"google.golang.org/grpc"
)

// Server holds the gRPC server implementation
// Note: This is a placeholder until protobuf code is generated with `make proto`
type Server struct {
	// After proto generation, embed the unimplemented server structs:
	// pb.UnimplementedOutageServiceServer
	// pb.UnimplementedNoteServiceServer
	// pb.UnimplementedTagServiceServer
	// pb.UnimplementedAlertServiceServer
	// pb.UnimplementedHealthServiceServer

	service *service.Service
}

// NewServer creates a new gRPC server
func NewServer(svc *service.Service) *Server {
	return &Server{
		service: svc,
	}
}

// RegisterServices registers all gRPC services with the gRPC server
// Note: This is a placeholder until protobuf code is generated
func (s *Server) RegisterServices(grpcServer *grpc.Server) {
	// After proto generation, uncomment and register services:
	// pb.RegisterOutageServiceServer(grpcServer, s)
	// pb.RegisterNoteServiceServer(grpcServer, s)
	// pb.RegisterTagServiceServer(grpcServer, s)
	// pb.RegisterAlertServiceServer(grpcServer, s)
	// pb.RegisterHealthServiceServer(grpcServer, s)
}

// TODO: After running `make proto`, implement the following gRPC service methods:
//
// OutageService:
//   - CreateOutage: Create a new outage with metadata and custom_fields
//   - GetOutage: Retrieve an outage with all fields including metadata
//   - ListOutages: List outages with pagination
//   - UpdateOutage: Update outage fields (use merge logic for metadata/custom_fields)
//   - DeleteOutage: Delete an outage
//
// NoteService:
//   - AddNote: Add a note with metadata and custom_fields
//   - GetNote: Retrieve a note
//   - ListNotesByOutage: List notes for an outage
//   - UpdateNote: Update a note (use merge logic for metadata/custom_fields)
//   - DeleteNote: Delete a note
//
// TagService:
//   - AddTag: Add a tag with custom_fields
//   - GetTag: Retrieve a tag
//   - ListTagsByOutage: List tags for an outage
//   - DeleteTag: Delete a tag
//   - SearchOutagesByTag: Search outages by tag key-value pair
//
// AlertService:
//   - ImportAlert: Import an alert with source_metadata, metadata, and custom_fields
//   - GetAlert: Retrieve an alert
//   - GetAlertByExternalID: Retrieve an alert by external ID and source
//   - ListAlertsByOutage: List alerts for an outage
//   - UpdateAlert: Update alert fields (use merge logic for metadata/custom_fields)
//
// HealthService:
//   - Check: Health check endpoint
//
// See internal/grpc/converters.go for converter function implementations.
