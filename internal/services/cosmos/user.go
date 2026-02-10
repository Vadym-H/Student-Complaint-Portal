package cosmos

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/models"
	"github.com/google/uuid"
)

// CreateUser inserts a user into the users container
func (s *Service) CreateUser(ctx context.Context, user *models.User) error {
	// Auto-generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Validate role
	if user.Role != models.RoleAdmin && user.Role != models.RoleStudent {
		s.log.Error("invalid user role", slog.String("role", user.Role))
		return ErrInvalidRole
	}

	containerClient, err := s.client.NewContainer(s.database, s.usersContainer)
	if err != nil {
		return err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	partitionKey := azcosmos.NewPartitionKeyString(user.ID)
	_, err = containerClient.CreateItem(ctx, partitionKey, userBytes, nil)
	return err
}

// GetUserByEmail retrieves a user from the users container by email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {

	containerClient, err := s.client.NewContainer(s.database, s.usersContainer)
	if err != nil {
		return nil, err
	}

	query := "SELECT * FROM c WHERE c.email = @email"

	queryOptions := &azcosmos.QueryOptions{
		QueryParameters: []azcosmos.QueryParameter{
			{
				Name:  "@email",
				Value: email,
			},
		},
	}

	pager := containerClient.NewQueryItemsPager(
		query,
		azcosmos.PartitionKey{}, // cross-partition
		queryOptions,
	)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Items {
			var user models.User
			if err := json.Unmarshal(item, &user); err != nil {
				return nil, err
			}

			return &user, nil
		}
	}

	return nil, nil // not found
}

// GetUserByID retrieves a user from the users container by ID
func (s *Service) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	containerClient, err := s.client.NewContainer(s.database, s.usersContainer)
	if err != nil {
		s.log.Error("failed to get users container", slog.String("error", err.Error()))
		return nil, err
	}

	partitionKey := azcosmos.NewPartitionKeyString(id)
	response, err := containerClient.ReadItem(ctx, partitionKey, id, nil)
	if err != nil {
		s.log.Error("failed to read user by ID", slog.String("userId", id), slog.String("error", err.Error()))
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal(response.Value, &user); err != nil {
		s.log.Error("failed to unmarshal user", slog.String("userId", id), slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Debug("user found by ID", slog.String("userId", id))
	return &user, nil
}

// GetUserByUsername retrieves a user from the users container by username
func (s *Service) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {

	containerClient, err := s.client.NewContainer(s.database, s.usersContainer)
	if err != nil {
		return nil, err
	}

	query := "SELECT * FROM c WHERE c.username = @username"

	queryOptions := &azcosmos.QueryOptions{
		QueryParameters: []azcosmos.QueryParameter{
			{
				Name:  "@username",
				Value: username,
			},
		},
	}

	pager := containerClient.NewQueryItemsPager(
		query,
		azcosmos.PartitionKey{}, // cross-partition
		queryOptions,
	)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Items {
			var user models.User
			if err := json.Unmarshal(item, &user); err != nil {
				return nil, err
			}

			return &user, nil
		}
	}

	return nil, nil // not found
}

// UpdateUser updates user information (name and/or username)
func (s *Service) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error) {
	containerClient, err := s.client.NewContainer(s.database, s.usersContainer)
	if err != nil {
		s.log.Error("failed to get users container", slog.String("error", err.Error()))
		return nil, err
	}

	// Get the current user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user for update", slog.String("userId", userID), slog.String("error", err.Error()))
		return nil, err
	}

	if user == nil {
		s.log.Debug("user not found for update", slog.String("userId", userID))
		return nil, ErrUserNotFound
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		user.Name = name
	}

	if username, ok := updates["username"].(string); ok && username != "" {
		// Check if new username is already taken by another user
		existingUser, err := s.GetUserByUsername(ctx, username)
		if err != nil {
			s.log.Error("failed to check existing username", slog.String("username", username), slog.String("error", err.Error()))
			return nil, err
		}
		if existingUser != nil && existingUser.ID != userID {
			s.log.Debug("username already taken", slog.String("username", username))
			return nil, ErrUsernameAlreadyExists
		}
		user.UserName = username
	}

	// Marshal the updated user
	userBytes, err := json.Marshal(user)
	if err != nil {
		s.log.Error("failed to marshal updated user", slog.String("userId", userID), slog.String("error", err.Error()))
		return nil, err
	}

	// Replace the item using the partition key
	partitionKey := azcosmos.NewPartitionKeyString(userID)
	_, err = containerClient.ReplaceItem(ctx, partitionKey, userID, userBytes, nil)
	if err != nil {
		s.log.Error("failed to update user in cosmos", slog.String("userId", userID), slog.String("error", err.Error()))
		return nil, err
	}

	s.log.Info("user updated successfully", slog.String("userId", userID))
	return user, nil
}
