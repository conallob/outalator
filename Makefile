.PHONY: help proto grpc-deps build run test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

grpc-deps: ## Install gRPC and protobuf tooling
	@echo "Installing protoc-gen-go..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	@echo "Installing protoc-gen-go-grpc..."
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	@echo "✓ gRPC dependencies installed"
	@echo ""
	@echo "NOTE: You also need to install protoc (Protocol Buffers compiler):"
	@echo "  - On macOS: brew install protobuf"
	@echo "  - On Ubuntu/Debian: sudo apt-get install -y protobuf-compiler"
	@echo "  - On other systems: https://grpc.io/docs/protoc-installation/"

proto: ## Generate Go code from proto files
	@./scripts/generate-proto.sh

build: ## Build the application
	@go build -o bin/outalator cmd/outalator/main.go
	@echo "✓ Built: bin/outalator"

run: ## Run the application
	@go run cmd/outalator/main.go

test: ## Run tests
	@go test -v ./...

clean: ## Clean build artifacts
	@rm -rf bin/
	@rm -f api/proto/*.pb.go
	@echo "✓ Cleaned build artifacts"
