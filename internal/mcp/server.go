package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/conall/outalator/internal/domain"
	"github.com/conall/outalator/internal/service"
	"github.com/google/uuid"
)

// Server represents the MCP server
type Server struct {
	service *service.Service
}

// NewServer creates a new MCP server instance
func NewServer(svc *service.Service) *Server {
	return &Server{
		service: svc,
	}
}

// Request represents an MCP request
type Request struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     interface{}     `json:"id"`
}

// Response represents an MCP response
type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *Error      `json:"error,omitempty"`
	ID     interface{} `json:"id"`
}

// Error represents an MCP error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Handle processes an MCP request
func (s *Server) Handle(ctx context.Context, req *Request) *Response {
	resp := &Response{ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = s.handleInitialize()
	case "tools/list":
		resp.Result = s.handleToolsList()
	case "tools/call":
		result, err := s.handleToolsCall(ctx, req.Params)
		if err != nil {
			resp.Error = &Error{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}
	default:
		resp.Error = &Error{
			Code:    -32601,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
	}

	return resp
}

// handleInitialize returns server capabilities
func (s *Server) handleInitialize() map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "outalator",
			"version": "1.0.0",
		},
	}
}

// handleToolsList returns available tools
func (s *Server) handleToolsList() map[string]interface{} {
	return map[string]interface{}{
		"tools": []map[string]interface{}{
			{
				"name":        "list_outages",
				"description": "List all outages with pagination support",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{
							"type":        "number",
							"description": "Maximum number of outages to return (default: 50, max: 100)",
						},
						"offset": map[string]interface{}{
							"type":        "number",
							"description": "Offset for pagination (default: 0)",
						},
					},
				},
			},
			{
				"name":        "get_outage",
				"description": "Get details of a specific outage by ID",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"outage_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID of the outage",
						},
					},
					"required": []string{"outage_id"},
				},
			},
			{
				"name":        "create_outage",
				"description": "Create a new outage entry",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Title of the outage",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Detailed description of the outage",
						},
						"severity": map[string]interface{}{
							"type":        "string",
							"description": "Severity level: critical, high, medium, low",
						},
					},
					"required": []string{"title", "description", "severity"},
				},
			},
			{
				"name":        "add_note",
				"description": "Add a note to an existing outage",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"outage_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID of the outage",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "Content of the note",
						},
						"author": map[string]interface{}{
							"type":        "string",
							"description": "Author of the note",
						},
						"format": map[string]interface{}{
							"type":        "string",
							"description": "Format of the note: plaintext or markdown (default: plaintext)",
						},
					},
					"required": []string{"outage_id", "content", "author"},
				},
			},
			{
				"name":        "update_outage",
				"description": "Update an existing outage",
				"inputSchema": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"outage_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID of the outage",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "New title (optional)",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "New description (optional)",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "New status: open, investigating, resolved, closed (optional)",
						},
						"severity": map[string]interface{}{
							"type":        "string",
							"description": "New severity: critical, high, medium, low (optional)",
						},
					},
					"required": []string{"outage_id"},
				},
			},
		},
	}
}

// handleToolsCall executes a tool
func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	switch callParams.Name {
	case "list_outages":
		return s.toolListOutages(ctx, callParams.Arguments)
	case "get_outage":
		return s.toolGetOutage(ctx, callParams.Arguments)
	case "create_outage":
		return s.toolCreateOutage(ctx, callParams.Arguments)
	case "add_note":
		return s.toolAddNote(ctx, callParams.Arguments)
	case "update_outage":
		return s.toolUpdateOutage(ctx, callParams.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", callParams.Name)
	}
}

// Tool implementations
func (s *Server) toolListOutages(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	limit := 50
	offset := 0

	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	outages, err := s.service.ListOutages(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Found %d outages", len(outages)),
			},
		},
		"outages": outages,
	}, nil
}

func (s *Server) toolGetOutage(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	outageIDStr, ok := args["outage_id"].(string)
	if !ok {
		return nil, fmt.Errorf("outage_id is required")
	}

	outageID, err := uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	outage, err := s.service.GetOutage(ctx, outageID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Outage: %s (Status: %s, Severity: %s)", outage.Title, outage.Status, outage.Severity),
			},
		},
		"outage": outage,
	}, nil
}

func (s *Server) toolCreateOutage(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	title, ok := args["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title is required")
	}

	description, ok := args["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description is required")
	}

	severity, ok := args["severity"].(string)
	if !ok {
		return nil, fmt.Errorf("severity is required")
	}

	req := domain.CreateOutageRequest{
		Title:       title,
		Description: description,
		Severity:    severity,
	}

	outage, err := s.service.CreateOutage(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Created outage: %s (ID: %s)", outage.Title, outage.ID),
			},
		},
		"outage": outage,
	}, nil
}

func (s *Server) toolAddNote(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	outageIDStr, ok := args["outage_id"].(string)
	if !ok {
		return nil, fmt.Errorf("outage_id is required")
	}

	outageID, err := uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}

	author, ok := args["author"].(string)
	if !ok {
		return nil, fmt.Errorf("author is required")
	}

	format := "plaintext"
	if f, ok := args["format"].(string); ok {
		format = f
	}

	req := domain.AddNoteRequest{
		Content: content,
		Format:  format,
		Author:  author,
	}

	note, err := s.service.AddNote(ctx, outageID, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Added note by %s to outage %s", note.Author, outageID),
			},
		},
		"note": note,
	}, nil
}

func (s *Server) toolUpdateOutage(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	outageIDStr, ok := args["outage_id"].(string)
	if !ok {
		return nil, fmt.Errorf("outage_id is required")
	}

	outageID, err := uuid.Parse(outageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid outage_id: %w", err)
	}

	req := domain.UpdateOutageRequest{}

	if title, ok := args["title"].(string); ok {
		req.Title = &title
	}
	if desc, ok := args["description"].(string); ok {
		req.Description = &desc
	}
	if status, ok := args["status"].(string); ok {
		req.Status = &status
	}
	if severity, ok := args["severity"].(string); ok {
		req.Severity = &severity
	}

	outage, err := s.service.UpdateOutage(ctx, outageID, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Updated outage: %s (Status: %s)", outage.Title, outage.Status),
			},
		},
		"outage": outage,
	}, nil
}

// ServeStdio serves the MCP protocol over stdin/stdout
func (s *Server) ServeStdio(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	decoder := json.NewDecoder(stdin)
	encoder := json.NewEncoder(stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		resp := s.Handle(ctx, &req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}
