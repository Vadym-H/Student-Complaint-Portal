package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/middleware"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/models"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services/cosmos"
	"github.com/google/uuid"
)

// ComplaintsHandler handles complaint-related requests
type ComplaintsHandler struct {
	cosmosService     *cosmos.Service
	serviceBusService *services.ServiceBusService
	log               *slog.Logger
}

// NewComplaintsHandler creates a new ComplaintsHandler
func NewComplaintsHandler(cosmosService *cosmos.Service, serviceBusService *services.ServiceBusService, log *slog.Logger) *ComplaintsHandler {
	const module = "complaintsHandler"
	log = log.With(
		slog.String("module", module),
	)
	return &ComplaintsHandler{
		cosmosService:     cosmosService,
		serviceBusService: serviceBusService,
		log:               log,
	}
}

// CreateComplaintRequest represents the request body for creating a complaint
type CreateComplaintRequest struct {
	Description string `json:"description"`
}

// CreateComplaint handles POST requests to create a new complaint
func (h *ComplaintsHandler) CreateComplaint(w http.ResponseWriter, r *http.Request) {
	// Get userId from context (set by auth middleware)
	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Parse JSON body
	var req CreateComplaintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", slog.String("userId", userId), slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate description
	if req.Description == "" {
		h.log.Error("description is empty", slog.String("userId", userId))
		http.Error(w, "Description cannot be empty", http.StatusBadRequest)
		return
	}

	// Create complaint
	complaint := &models.Complaint{
		ID:          uuid.New().String(),
		UserID:      userId,
		Description: req.Description,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
	}

	// Save complaint to Cosmos DB
	if err := h.cosmosService.CreateComplaint(r.Context(), complaint); err != nil {
		h.log.Error("failed to create complaint", slog.String("userId", userId), slog.String("complaintId", complaint.ID), slog.String("error", err.Error()))
		http.Error(w, "Failed to create complaint", http.StatusInternalServerError)
		return
	}

	// Send complaint ID to Service Bus queue
	if err := h.serviceBusService.SendMessage(r.Context(), "new-complaints", complaint.ID); err != nil {
		h.log.Error("failed to send complaint to service bus", slog.String("userId", userId), slog.String("complaintId", complaint.ID), slog.String("error", err.Error()))
		http.Error(w, "Failed to queue complaint", http.StatusInternalServerError)
		return
	}

	// Log successful complaint creation
	h.log.Info("complaint created successfully", slog.String("userId", userId), slog.String("complaintId", complaint.ID))

	// Return created complaint as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(complaint); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("complaintId", complaint.ID), slog.String("error", err.Error()))
	}
}

// GetComplaints handles GET requests to retrieve complaints
func (h *ComplaintsHandler) GetComplaints(w http.ResponseWriter, r *http.Request) {
	// Get userId and role from context (set by auth middleware)
	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	role, ok := middleware.GetRole(r.Context())
	if !ok {
		h.log.Error("failed to get role from context", slog.String("userId", userId), slog.String("path", r.URL.Path))
		http.Error(w, "Role not found in context", http.StatusInternalServerError)
		return
	}

	// Read query parameters
	status := r.URL.Query().Get("status")
	id := r.URL.Query().Get("id")

	h.log.Info("getting complaints", slog.String("userId", userId), slog.String("role", role), slog.String("status", status), slog.String("complaintId", id))

	var complaints []models.Complaint
	var err error

	if role == models.RoleAdmin {
		// Admin can view specific complaint by ID or all complaints filtered by status
		if id != "" {
			// Return single complaint by ID
			complaint, err := h.cosmosService.GetComplaintByID(r.Context(), id)
			if err != nil {
				h.log.Error("failed to get complaint by ID", slog.String("adminId", userId), slog.String("complaintId", id), slog.String("error", err.Error()))
				http.Error(w, "Failed to retrieve complaint", http.StatusInternalServerError)
				return
			}

			if complaint == nil {
				h.log.Warn("complaint not found", slog.String("adminId", userId), slog.String("complaintId", id))
				http.Error(w, "Complaint not found", http.StatusNotFound)
				return
			}

			complaints = []models.Complaint{*complaint}
			h.log.Info("complaint retrieved by ID", slog.String("adminId", userId), slog.String("complaintId", id))
		} else {
			// Return all complaints filtered by status (if provided)
			// For admin viewing all complaints, we need to query without userId filter
			// This would require a different cosmos query - for now we'll get by status across all
			h.log.Warn("admin requesting all complaints - this requires cross-partition query", slog.String("adminId", userId))
			// Note: This would need a separate cosmos method or a cross-partition query
			// For simplicity, returning empty for now - in production you'd implement this differently
			complaints = []models.Complaint{}
		}
	} else if role == models.RoleStudent {
		// Student can only view their own complaints, optionally filtered by status
		complaints, err = h.cosmosService.GetComplaints(r.Context(), userId, status)
		if err != nil {
			h.log.Error("failed to get complaints", slog.String("userId", userId), slog.String("status", status), slog.String("error", err.Error()))
			http.Error(w, "Failed to retrieve complaints", http.StatusInternalServerError)
			return
		}
		h.log.Info("complaints retrieved for student", slog.String("userId", userId), slog.String("status", status), slog.Int("count", len(complaints)))
	} else {
		h.log.Error("unknown role", slog.String("userId", userId), slog.String("role", role))
		http.Error(w, "Unknown role", http.StatusForbidden)
		return
	}

	// Return complaints as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(complaints); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("error", err.Error()))
	}
}

// UpdateComplaintRequest represents the request body for updating a complaint
type UpdateComplaintRequest struct {
	Status string `json:"status"`
}

// UpdateComplaint handles PUT requests to update a complaint (admin-only)
func (h *ComplaintsHandler) UpdateComplaint(w http.ResponseWriter, r *http.Request) {
	// Get adminId from context (set by auth middleware)
	adminId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Get complaint ID from URL parameter
	complaintId := r.PathValue("id")
	if complaintId == "" {
		h.log.Error("complaint id not provided in URL", slog.String("adminId", adminId))
		http.Error(w, "Complaint ID required", http.StatusBadRequest)
		return
	}

	// Parse JSON body
	var req UpdateComplaintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request body", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	if req.Status == "" {
		h.log.Error("status is empty", slog.String("adminId", adminId), slog.String("complaintId", complaintId))
		http.Error(w, "Status cannot be empty", http.StatusBadRequest)
		return
	}

	// Validate allowed status transitions
	allowedStatuses := map[string]bool{
		models.StatusPending:  true,
		models.StatusApproved: true,
		models.StatusRejected: true,
	}
	if !allowedStatuses[req.Status] {
		h.log.Error("invalid status value", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("status", req.Status))
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	// Update complaint status in Cosmos DB
	if err := h.cosmosService.UpdateComplaintStatus(r.Context(), complaintId, req.Status); err != nil {
		h.log.Error("failed to update complaint status", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("error", err.Error()))
		http.Error(w, "Failed to update complaint", http.StatusInternalServerError)
		return
	}

	// Send complaint ID to Service Bus queue
	if err := h.serviceBusService.SendMessage(r.Context(), "complaint-status-changed", complaintId); err != nil {
		h.log.Error("failed to send complaint to service bus", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("error", err.Error()))
		http.Error(w, "Failed to queue status change notification", http.StatusInternalServerError)
		return
	}

	// Log successful status change
	h.log.Info("complaint status updated", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("newStatus", req.Status))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message":     "Complaint status updated successfully",
		"complaintId": complaintId,
		"status":      req.Status,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", slog.String("adminId", adminId), slog.String("complaintId", complaintId), slog.String("error", err.Error()))
	}
}
