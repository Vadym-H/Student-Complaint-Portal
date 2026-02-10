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
	Role     string `json:"role"`
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
		Role:     user.Role,
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("error", err.Error()))
	}
}

// UpdateUserProfileRequest represents the update profile request body
type UpdateUserProfileRequest struct {
	Name     string `json:"name,omitempty"`
	UserName string `json:"username,omitempty"`
}

// UpdateUserProfile handles PUT requests to update the current user's profile
// @Summary Update current user's profile
// @Description Update the name and/or username of the authenticated user
// @Tags users
// @Security Bearer
// @Produce json
// @Param request body UpdateUserProfileRequest true "Update profile request"
// @Success 200 {object} UserInfoResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "Conflict - Username already exists"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/users/me [put]
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get userId from context (set by auth middleware)
	userId, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.log.Error("failed to get userId from context", slog.String("path", r.URL.Path))
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req UpdateUserProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debug("failed to parse update profile request", slog.String("userId", userId), slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate at least one field is provided
	if req.Name == "" && req.UserName == "" {
		h.log.Debug("update profile request with no fields", slog.String("userId", userId))
		http.Error(w, "At least one field (name or username) must be provided", http.StatusBadRequest)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.UserName != "" {
		updates["username"] = req.UserName
	}

	// Update user in database
	user, err := h.cosmosService.UpdateUser(r.Context(), userId, updates)
	if err != nil {
		if err.Error() == cosmos.ErrUsernameAlreadyExists.Error() {
			h.log.Debug("username already exists", slog.String("userId", userId), slog.String("username", req.UserName))
			http.Error(w, cosmos.ErrUsernameAlreadyExists.Error(), http.StatusConflict)
			return
		}
		if err.Error() == cosmos.ErrUserNotFound.Error() {
			h.log.Error("user not found for update", slog.String("userId", userId))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.log.Error("failed to update user", slog.String("userId", userId), slog.String("error", err.Error()))
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Build response
	response := UserInfoResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		UserName: user.UserName,
		Role:     user.Role,
	}

	h.log.Info("user profile updated successfully", slog.String("userId", userId))

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", slog.String("userId", userId), slog.String("error", err.Error()))
	}
}
