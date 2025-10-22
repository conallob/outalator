#!/bin/bash

set -e

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed."
    echo "Please install protoc (Protocol Buffers compiler):"
    echo "  - On macOS: brew install protobuf"
    echo "  - On Ubuntu/Debian: apt-get install -y protobuf-compiler"
    echo "  - On other systems: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
fi

# Create output directory
mkdir -p api/proto/v1

# Generate Go code from proto files
echo "Generating Go code from proto files..."
protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    api/proto/outalator.proto

echo "✓ Proto files generated successfully!"
echo "  Generated files:"
echo "    - api/proto/outalator.pb.go (messages)"
echo "    - api/proto/outalator_grpc.pb.go (services)"
