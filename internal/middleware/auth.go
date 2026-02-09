package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Context keys for storing user information in request context
type contextKey string

const (
	userIDKey contextKey = "userId"
	roleKey   contextKey = "role"
)

// Claims Custom claims structure for JWT
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a JWT token with the provided user information
func GenerateJWT(userId, email, role, secret string) (string, error) {
	// Create claims with user information and expiration time
	claims := Claims{
		UserID: userId,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RequireAuth middleware validates JWT token and adds user info to context
func RequireAuth(secret string, log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// First, try to get token from cookie
			cookie, err := r.Cookie("auth_token")
			if err == nil && cookie.Value != "" {
				tokenString = cookie.Value
				log.Debug("token found in cookie", slog.String("path", r.URL.Path))
			} else {
				// Fall back to Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					log.Debug("no token found in cookie or authorization header", slog.String("path", r.URL.Path))
					http.Error(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				// Extract token from "Bearer <token>" format
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					log.Debug("invalid authorization header format", slog.String("path", r.URL.Path))
					http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
					return
				}

				tokenString = parts[1]
				log.Debug("token found in authorization header", slog.String("path", r.URL.Path))
			}

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				return []byte(secret), nil
			})

			if err != nil {
				log.Debug("failed to parse JWT token", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims
			claims, ok := token.Claims.(*Claims)
			if !ok || !token.Valid {
				log.Debug("invalid token claims", slog.String("path", r.URL.Path))
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add userId and role to request context
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin middleware checks if user has admin role
func RequireAdmin(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context
			role, ok := r.Context().Value(roleKey).(string)
			if !ok {
				log.Debug("role not found in context", slog.String("path", r.URL.Path))
				http.Error(w, "Unauthorized: role not found", http.StatusForbidden)
				return
			}

			// Check if role is admin
			if role != "admin" {
				userId, _ := GetUserID(r.Context())
				log.Debug("non-admin user attempted admin action", slog.String("userId", userId), slog.String("role", role), slog.String("path", r.URL.Path))
				http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions to extract user info from context

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userId, ok := ctx.Value(userIDKey).(string)
	return userId, ok
}

// GetRole extracts the role from the request context
// Currently unused, but will be needed for complaint handlers to implement role-based logic
// (e.g., students can only view their own complaints, admins can view all)
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}
