package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/middleware"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services/cosmos"
)

// UserHandler handles user-related requests
type UserHandler struct {
	cosmosService *cosmos.Service
	log           *slog.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(cosmosService *cosmos.Service, log *slog.Logger) *UserHandler {
	const module = "userHandler"
	log = log.With(
		slog.String("module", module),
	)
	return &UserHandler{
		cosmosService: cosmosService,
		log:           log,
	}
}

// UserInfoResponse represents the user info response body
type UserInfoResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	UserName string `json:"username"`
}

// GetUserInfo handles GET requests to retrieve the current user's information
// @Summary Get current user information
// @Description Get the email, name, and username of the authenticated user
// @Tags users
// @Security Bearer
// @Produce json
// @Success 200 {object} UserInfoResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/users/me [get]
func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// Get userId from context (set by auth middleware)
	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Fetch user from Cosmos DB
	user, err := h.cosmosService.GetUserByID(r.Context(), userId)
	if err != nil {
		h.log.Error("failed to get user from cosmos DB", slog.String("userId", userId), slog.String("error", err.Error()))
		http.Error(w, "Failed to retrieve user information", http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.log.Error("user not found", slog.String("userId", userId))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Build response
	response := UserInfoResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		UserName: user.UserName,
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("error", err.Error()))
	}
}
