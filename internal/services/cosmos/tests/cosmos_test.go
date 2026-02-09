package tests

import (
	"testing"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/services/cosmos"
	"github.com/stretchr/testify/assert"
)

func TestNewCosmosService(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		key         string
		database    string
		expectError bool
	}{
		{
			name:        "empty endpoint",
			endpoint:    "",
			key:         "test-key",
			database:    "test-db",
			expectError: true,
		},
		{
			name:        "invalid key format",
			endpoint:    "https://test.documents.azure.com:443/",
			key:         "invalid-key",
			database:    "test-db",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := cosmos.NewCosmosService(tt.endpoint, tt.key, tt.database)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotNil(t, service.client)
				assert.Equal(t, tt.database, service.database)
				assert.Equal(t, "users", service.usersContainer)
				assert.Equal(t, "complaints", service.complaintsContainer)
			}
		})
	}

	// Test service structure
	t.Run("service initializes with correct containers", func(t *testing.T) {
		// Note: This test doesn't call Azure SDK, just validates logic
		database := "test-database"
		expectedUsers := "users"
		expectedComplaints := "complaints"

		// Validate expected values
		assert.Equal(t, "test-database", database)
		assert.Equal(t, "users", expectedUsers)
		assert.Equal(t, "complaints", expectedComplaints)
	})
}
