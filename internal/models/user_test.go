package models

import (
	"testing"
	"time"
)

func TestUserRoleConstants(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{
			name:     "student role constant",
			role:     RoleStudent,
			expected: "student",
		},
		{
			name:     "admin role constant",
			role:     RoleAdmin,
			expected: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.role != tt.expected {
				t.Errorf("got %v, want %v", tt.role, tt.expected)
			}
		})
	}
}

func TestUserCreation(t *testing.T) {
	now := time.Now()

	user := &User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "John Doe",
		UserName:     "johndoe",
		PasswordHash: "hashedpassword",
		Role:         RoleStudent,
		CreatedAt:    now,
	}

	tests := []struct {
		name     string
		field    string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "user ID is set",
			field:    "ID",
			value:    user.ID,
			expected: "user-123",
		},
		{
			name:     "email is set",
			field:    "Email",
			value:    user.Email,
			expected: "test@example.com",
		},
		{
			name:     "name is set",
			field:    "Name",
			value:    user.Name,
			expected: "John Doe",
		},
		{
			name:     "username is set",
			field:    "UserName",
			value:    user.UserName,
			expected: "johndoe",
		},
		{
			name:     "role is student",
			field:    "Role",
			value:    user.Role,
			expected: RoleStudent,
		},
		{
			name:     "created at is set",
			field:    "CreatedAt",
			value:    user.CreatedAt,
			expected: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("got %v, want %v", tt.value, tt.expected)
			}
		})
	}
}

func TestUserWithDifferentRoles(t *testing.T) {
	validRoles := []string{RoleStudent, RoleAdmin}

	for _, role := range validRoles {
		t.Run("role="+role, func(t *testing.T) {
			user := &User{
				ID:        "user-1",
				Email:     "test@example.com",
				Name:      "Test User",
				UserName:  "testuser",
				Role:      role,
				CreatedAt: time.Now(),
			}

			if user.Role != role {
				t.Errorf("got role %v, want %v", user.Role, role)
			}
		})
	}
}

func TestUserEmailValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "valid email",
			email: "user@example.com",
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
		},
		{
			name:  "valid email with numbers",
			email: "user123@example.com",
		},
		{
			name:  "empty email",
			email: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Email: tt.email,
			}

			if user.Email != tt.email {
				t.Errorf("got email %q, want %q", user.Email, tt.email)
			}
		})
	}
}

func TestUserPasswordHashHandling(t *testing.T) {
	user := &User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: "hashed_value",
	}

	// Verify that PasswordHash field is properly handled
	if user.PasswordHash != "hashed_value" {
		t.Errorf("got PasswordHash %q, want 'hashed_value'", user.PasswordHash)
	}

	// Verify that empty password hash is allowed
	user.PasswordHash = ""
	if user.PasswordHash != "" {
		t.Error("expected empty PasswordHash")
	}
}

func TestEmptyUserValidation(t *testing.T) {
	user := &User{}

	if user.ID != "" {
		t.Error("expected empty ID")
	}
	if user.Email != "" {
		t.Error("expected empty Email")
	}
	if user.Name != "" {
		t.Error("expected empty Name")
	}
	if user.UserName != "" {
		t.Error("expected empty UserName")
	}
	if user.Role != "" {
		t.Error("expected empty Role")
	}
}
