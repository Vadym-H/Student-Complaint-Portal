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
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(complaint); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("complaintId", complaint.ID), slog.String("error", err.Error()))
	}
}

// GetComplaints handles GET requests to retrieve complaints for the authenticated user
func (h *ComplaintsHandler) GetComplaints(w http.ResponseWriter, r *http.Request) {
	// Get userId from context (set by auth middleware)
	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	// Read query parameter for status filter
	status := r.URL.Query().Get("status")

	h.log.Info("getting user complaints", slog.String("userId", userId), slog.String("status", status))

	// Get complaints for this specific user only, optionally filtered by status
	complaints, err := h.cosmosService.GetComplaints(r.Context(), userId, status)
	if err != nil {
		h.log.Error("failed to get complaints", slog.String("userId", userId), slog.String("status", status), slog.String("error", err.Error()))
		http.Error(w, "Failed to retrieve complaints", http.StatusInternalServerError)
		return
	}
	
	h.log.Info("complaints retrieved for user", slog.String("userId", userId), slog.String("status", status), slog.Int("count", len(complaints)))

	// Return complaints as JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
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

// GetAllComplaintsAdmin handles GET requests to retrieve all complaints (admin-only)
// @Summary Get all complaints (admin)
// @Description Get all complaints, optionally filtered by status
// @Tags admin
// @Security Bearer
// @Produce json
// @Success 200 {array} models.Complaint
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/admin/complaints [get]
func (h *ComplaintsHandler) GetAllComplaintsAdmin(w http.ResponseWriter, r *http.Request) {
	adminId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	role, ok := middleware.GetRole(r.Context())
	if !ok {
		h.log.Error("failed to get role from context", slog.String("adminId", adminId), slog.String("path", r.URL.Path))
		http.Error(w, "Role not found in context", http.StatusInternalServerError)
		return
	}
	if role != models.RoleAdmin {
		h.log.Warn("non-admin attempted admin complaints", slog.String("userId", adminId), slog.String("role", role))
		http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
		return
	}

	status := r.URL.Query().Get("status")
	h.log.Info("admin getting all complaints", slog.String("adminId", adminId), slog.String("status", status))

	complaints, err := h.cosmosService.GetAllComplaints(r.Context(), status)
	if err != nil {
		h.log.Error("failed to get all complaints", slog.String("adminId", adminId), slog.String("status", status), slog.String("error", err.Error()))
		http.Error(w, "Failed to retrieve complaints", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(complaints); err != nil {
		h.log.Error("failed to encode response", slog.String("adminId", adminId), slog.String("error", err.Error()))
	}
}
