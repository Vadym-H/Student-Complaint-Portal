package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/middleware"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/models"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services/cosmos"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	cosmosService *cosmos.Service
	jwtSecret     string
	log           *slog.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(cosmosService *cosmos.Service, jwtSecret string, log *slog.Logger) *AuthHandler {
	const module = "authHandler"
	log = log.With(
		slog.String("module", module),
	)
	return &AuthHandler{
		cosmosService: cosmosService,
		jwtSecret:     jwtSecret,
		log:           log,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	UserName string `json:"username"`
	Name     string `json:"name"`
	Token    string `json:"token"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.UserName == "" || req.Name == "" {
		http.Error(w, "Email, username, name, and password are required", http.StatusBadRequest)
		return
	}

	// Check if user with this email already exists
	existingUser, err := h.cosmosService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.log.Error("failed to check existing user by email", slog.String("email", req.Email), slog.String("error", err.Error()))
		http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		h.log.Debug("registration attempt with existing email", slog.String("email", req.Email))
		http.Error(w, cosmos.ErrEmailAlreadyExists.Error(), http.StatusConflict)
		return
	}

	// Check if user with this username already exists
	existingUserByUsername, err := h.cosmosService.GetUserByUsername(r.Context(), req.UserName)
	if err != nil {
		h.log.Error("failed to check existing user by username", slog.String("username", req.UserName), slog.String("error", err.Error()))
		http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
		return
	}
	if existingUserByUsername != nil {
		h.log.Debug("registration attempt with existing username", slog.String("username", req.UserName))
		http.Error(w, cosmos.ErrUsernameAlreadyExists.Error(), http.StatusConflict)
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("failed to hash password", slog.String("error", err.Error()))
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		UserName:     req.UserName,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	// Save user to database
	if err := h.cosmosService.CreateUser(r.Context(), user); err != nil {
		h.log.Error("failed to create user", slog.String("userId", user.ID), slog.String("error", err.Error()))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		h.log.Error("failed to generate JWT token", slog.String("userId", user.ID), slog.String("error", err.Error()))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return response
	response := RegisterResponse{
		ID:       user.ID,
		Email:    user.Email,
		UserName: user.UserName,
		Name:     user.Name,
		Token:    token,
	}

	h.log.Info("user registered successfully", slog.String("userId", user.ID), slog.String("email", user.Email), slog.String("username", user.UserName))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.log.Error("failed to encode response", slog.String("userId", user.ID), slog.String("error", err.Error()))
		return
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Debug("failed to parse login request", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		h.log.Debug("login request missing required fields")
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.cosmosService.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.log.Error("failed to retrieve user", slog.String("email", req.Email), slog.String("error", err.Error()))
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		h.log.Debug("login attempt with non-existent email", slog.String("email", req.Email))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.log.Debug("login attempt with invalid password", slog.String("email", req.Email))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		h.log.Error("failed to generate JWT token", slog.String("userId", user.ID), slog.String("error", err.Error()))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set HTTP-only cookie with token
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	h.log.Info("user logged in successfully", slog.String("userId", user.ID), slog.String("email", user.Email), slog.String("role", user.Role))

	// Return response (also include token for backward compatibility)
	response := LoginResponse{
		Token: token,
		Role:  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.log.Error("Failed to encode response")
		return
	}
}

// Logout handles user logout by clearing the auth cookie
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie by setting MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Delete cookie
	})

	h.log.Info("user logged out successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"message":"Logged out successfully"}`))
	if err != nil {
		h.log.Warn("Failed to write")
		return
	}
}
