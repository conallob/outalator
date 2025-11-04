package grpc

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note: This file contains placeholder converter functions for gRPC.
// These functions will be implemented after running `make proto` to generate the protobuf code.
//
// To generate protobuf code:
//   1. Install protoc: brew install protobuf (macOS) or apt-get install protobuf-compiler (Ubuntu)
//   2. Install Go plugins: make grpc-deps
//   3. Generate code: make proto
//
// After generation, uncomment and complete the conversion functions in this file.
// See docs/GRPC_SETUP.md for detailed setup instructions.

// Helper functions for converting between domain models and protobuf messages

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

// TODO: After running `make proto`, implement the following converter functions:
// - OutageDomainToProto: Convert domain.Outage to pb.Outage (include metadata and custom_fields)
// - AlertDomainToProto: Convert domain.Alert to pb.Alert (include source_metadata, metadata, and custom_fields)
// - NoteDomainToProto: Convert domain.Note to pb.Note (include metadata and custom_fields)
// - TagDomainToProto: Convert domain.Tag to pb.Tag (include custom_fields)
// - CreateOutageRequestProtoToDomain: Convert pb.CreateOutageRequest to domain.CreateOutageRequest
// - And corresponding reverse converters for all domain types
//
// Important: Ensure all new fields (metadata, custom_fields, source_metadata) are properly converted
// between protobuf Struct/map types and Go map[string]string/map[string]any types.
