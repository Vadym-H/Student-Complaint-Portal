package models

import (
	"testing"
	"time"
)

func TestComplaintStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "pending status constant",
			status:   StatusPending,
			expected: "pending",
		},
		{
			name:     "approved status constant",
			status:   StatusApproved,
			expected: "approved",
		},
		{
			name:     "rejected status constant",
			status:   StatusRejected,
			expected: "rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("got %v, want %v", tt.status, tt.expected)
			}
		})
	}
}

func TestComplaintCreation(t *testing.T) {
	now := time.Now()

	complaint := &Complaint{
		ID:          "test-id-123",
		UserID:      "user-456",
		Description: "This is a test complaint",
		Status:      StatusPending,
		CreatedAt:   now,
	}

	tests := []struct {
		name     string
		field    string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "complaint ID is set",
			field:    "ID",
			value:    complaint.ID,
			expected: "test-id-123",
		},
		{
			name:     "user ID is set",
			field:    "UserID",
			value:    complaint.UserID,
			expected: "user-456",
		},
		{
			name:     "description is set",
			field:    "Description",
			value:    complaint.Description,
			expected: "This is a test complaint",
		},
		{
			name:     "status is pending",
			field:    "Status",
			value:    complaint.Status,
			expected: StatusPending,
		},
		{
			name:     "created at is set",
			field:    "CreatedAt",
			value:    complaint.CreatedAt,
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

func TestEmptyComplaintValidation(t *testing.T) {
	complaint := &Complaint{}

	if complaint.ID != "" {
		t.Error("expected empty ID")
	}
	if complaint.UserID != "" {
		t.Error("expected empty UserID")
	}
	if complaint.Description != "" {
		t.Error("expected empty Description")
	}
	if complaint.Status != "" {
		t.Error("expected empty Status")
	}
}

func TestComplaintWithDifferentStatuses(t *testing.T) {
	validStatuses := []string{StatusPending, StatusApproved, StatusRejected}

	for _, status := range validStatuses {
		t.Run("status="+status, func(t *testing.T) {
			complaint := &Complaint{
				ID:          "id-1",
				UserID:      "user-1",
				Description: "Test complaint",
				Status:      status,
				CreatedAt:   time.Now(),
			}

			if complaint.Status != status {
				t.Errorf("got status %v, want %v", complaint.Status, status)
			}
		})
	}
}
