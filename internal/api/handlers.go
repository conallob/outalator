package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/conall/outalator/internal/auth"
	"github.com/conall/outalator/internal/domain"
	"github.com/conall/outalator/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all HTTP routes
func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Outage routes
	r.HandleFunc("/api/v1/outages", h.CreateOutage).Methods("POST")
	r.HandleFunc("/api/v1/outages", h.ListOutages).Methods("GET")
	r.HandleFunc("/api/v1/outages/{id}", h.GetOutage).Methods("GET")
	r.HandleFunc("/api/v1/outages/{id}", h.UpdateOutage).Methods("PATCH")

	// Note routes
	r.HandleFunc("/api/v1/outages/{id}/notes", h.AddNote).Methods("POST")

	// Tag routes
	r.HandleFunc("/api/v1/outages/{id}/tags", h.AddTag).Methods("POST")
	r.HandleFunc("/api/v1/tags/search", h.SearchByTag).Methods("GET")

	// Alert routes
	r.HandleFunc("/api/v1/alerts/import", h.ImportAlert).Methods("POST")

	// Health check
	r.HandleFunc("/health", h.Health).Methods("GET")
}

// CreateOutage handles POST /api/v1/outages
func (h *Handler) CreateOutage(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateOutageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	outage, err := h.service.CreateOutage(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, outage)
}

// ListOutages handles GET /api/v1/outages
func (h *Handler) ListOutages(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}

	outages, err := h.service.ListOutages(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"outages": outages,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetOutage handles GET /api/v1/outages/{id}
func (h *Handler) GetOutage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid outage ID")
		return
	}

	outage, err := h.service.GetOutage(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Outage not found")
		return
	}

	respondJSON(w, http.StatusOK, outage)
}

// UpdateOutage handles PATCH /api/v1/outages/{id}
func (h *Handler) UpdateOutage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid outage ID")
		return
	}

	var req domain.UpdateOutageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	outage, err := h.service.UpdateOutage(r.Context(), id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, outage)
}

// AddNote handles POST /api/v1/outages/{id}/notes
func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid outage ID")
		return
	}

	// Get authenticated user
	user, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		respondError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req domain.AddNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Override author with authenticated user's email
	req.Author = user.Email

	note, err := h.service.AddNote(r.Context(), id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, note)
}

// AddTag handles POST /api/v1/outages/{id}/tags
func (h *Handler) AddTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid outage ID")
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tag, err := h.service.AddTag(r.Context(), id, req.Key, req.Value)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, tag)
}

// SearchByTag handles GET /api/v1/tags/search?key=...&value=...
func (h *Handler) SearchByTag(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" || value == "" {
		respondError(w, http.StatusBadRequest, "Both key and value parameters are required")
		return
	}

	outages, err := h.service.FindOutagesByTag(r.Context(), key, value)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"outages": outages,
	})
}

// ImportAlert handles POST /api/v1/alerts/import
func (h *Handler) ImportAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source     string     `json:"source"`
		ExternalID string     `json:"external_id"`
		OutageID   *uuid.UUID `json:"outage_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	alert, err := h.service.ImportAlert(r.Context(), req.Source, req.ExternalID, req.OutageID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, alert)
}

// Health handles GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}
