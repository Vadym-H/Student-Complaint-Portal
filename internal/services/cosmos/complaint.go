package cosmos

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/models"
	"github.com/google/uuid"
)

// CreateComplaint inserts a complaint into the complaints container
func (s *Service) CreateComplaint(ctx context.Context, complaint *models.Complaint) error {
	// Auto-generate ID if not provided
	if complaint.ID == "" {
		complaint.ID = uuid.New().String()
	}

	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		return err
	}

	complaintBytes, err := json.Marshal(complaint)
	if err != nil {
		return err
	}

	// Use UserID as partition keys (matches /userId in Terraform config)
	partitionKey := azcosmos.NewPartitionKeyString(complaint.UserID)
	_, err = containerClient.CreateItem(ctx, partitionKey, complaintBytes, nil)
	return err
}

// GetComplaints retrieves complaints from the complaints container by userId and optionally filters by status
func (s *Service) GetComplaints(ctx context.Context, userId, status string) ([]models.Complaint, error) {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		s.log.Error("failed to get complaints container", slog.String("error", err.Error()))
		return nil, err
	}

	var query string
	var queryOptions *azcosmos.QueryOptions

	if status != "" {
		// Filter by both userId and status
		query = "SELECT * FROM c WHERE c.userId = @userId AND c.status = @status"
		queryOptions = &azcosmos.QueryOptions{
			QueryParameters: []azcosmos.QueryParameter{
				{Name: "@userId", Value: userId},
				{Name: "@status", Value: status},
			},
		}
	} else {
		// Filter by userId only
		query = "SELECT * FROM c WHERE c.userId = @userId"
		queryOptions = &azcosmos.QueryOptions{
			QueryParameters: []azcosmos.QueryParameter{
				{Name: "@userId", Value: userId},
			},
		}
	}

	// Use partition key for efficient query
	partitionKey := azcosmos.NewPartitionKeyString(userId)
	pager := containerClient.NewQueryItemsPager(query, partitionKey, queryOptions)

	var complaints []models.Complaint
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			s.log.Error("failed to query complaints", slog.String("userId", userId), slog.String("status", status), slog.String("error", err.Error()))
			return nil, err
		}

		for _, item := range page.Items {
			var complaint models.Complaint
			if err := json.Unmarshal(item, &complaint); err != nil {
				s.log.Error("failed to unmarshal complaint", slog.String("error", err.Error()))
				return nil, err
			}
			complaints = append(complaints, complaint)
		}
	}

	s.log.Debug("complaints retrieved", slog.String("userId", userId), slog.String("status", status), slog.Int("count", len(complaints)))
	return complaints, nil
}

// UpdateComplaintStatus updates the status of a complaint by ID
func (s *Service) UpdateComplaintStatus(ctx context.Context, id, status string) error {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		s.log.Error("failed to get complaints container", slog.String("error", err.Error()))
		return err
	}

	// First, find the complaint to get the partition key (userId)
	query := "SELECT * FROM c WHERE c.id = @id"
	queryOptions := &azcosmos.QueryOptions{
		QueryParameters: []azcosmos.QueryParameter{
			{Name: "@id", Value: id},
		},
	}

	pager := containerClient.NewQueryItemsPager(query, azcosmos.PartitionKey{}, queryOptions)

	var complaint *models.Complaint
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			s.log.Error("failed to query complaint for update", slog.String("complaintId", id), slog.String("error", err.Error()))
			return err
		}

		for _, item := range page.Items {
			var c models.Complaint
			if err := json.Unmarshal(item, &c); err != nil {
				s.log.Error("failed to unmarshal complaint", slog.String("complaintId", id), slog.String("error", err.Error()))
				return err
			}
			complaint = &c
			break
		}
		if complaint != nil {
			break
		}
	}

	if complaint == nil {
		s.log.Debug("complaint not found for update", slog.String("complaintId", id))
		return nil // complaint not found
	}

	// Update the status
	oldStatus := complaint.Status
	complaint.Status = status

	// Marshal the updated complaint
	complaintBytes, err := json.Marshal(complaint)
	if err != nil {
		s.log.Error("failed to marshal updated complaint", slog.String("complaintId", id), slog.String("error", err.Error()))
		return err
	}

	// Replace the item using the partition key
	partitionKey := azcosmos.NewPartitionKeyString(complaint.UserID)
	_, err = containerClient.ReplaceItem(ctx, partitionKey, id, complaintBytes, nil)
	if err != nil {
		s.log.Error("failed to update complaint status in cosmos", slog.String("complaintId", id), slog.String("error", err.Error()))
		return err
	}

	s.log.Info("complaint status updated", slog.String("complaintId", id), slog.String("oldStatus", oldStatus), slog.String("newStatus", status))
	return nil
}

// GetComplaintByID retrieves a single complaint by its ID
func (s *Service) GetComplaintByID(ctx context.Context, id string) (*models.Complaint, error) {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		s.log.Error("failed to get complaints container", slog.String("error", err.Error()))
		return nil, err
	}

	query := "SELECT * FROM c WHERE c.id = @id"
	queryOptions := &azcosmos.QueryOptions{
		QueryParameters: []azcosmos.QueryParameter{
			{Name: "@id", Value: id},
		},
	}

	pager := containerClient.NewQueryItemsPager(query, azcosmos.PartitionKey{}, queryOptions)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			s.log.Error("failed to query complaint by ID", slog.String("complaintId", id), slog.String("error", err.Error()))
			return nil, err
		}

		for _, item := range page.Items {
			var complaint models.Complaint
			if err := json.Unmarshal(item, &complaint); err != nil {
				s.log.Error("failed to unmarshal complaint", slog.String("complaintId", id), slog.String("error", err.Error()))
				return nil, err
			}
			s.log.Debug("complaint retrieved by ID", slog.String("complaintId", id))
			return &complaint, nil
		}
	}

	s.log.Debug("complaint not found by ID", slog.String("complaintId", id))
	return nil, nil
}

// GetAllComplaints retrieves all complaints, optionally filtered by status.
func (s *Service) GetAllComplaints(ctx context.Context, status string) ([]models.Complaint, error) {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		s.log.Error("failed to get complaints container", slog.String("error", err.Error()))
		return nil, err
	}

	query := "SELECT * FROM c"
	var queryOptions *azcosmos.QueryOptions
	if status != "" {
		query = "SELECT * FROM c WHERE c.status = @status"
		queryOptions = &azcosmos.QueryOptions{
			QueryParameters: []azcosmos.QueryParameter{
				{Name: "@status", Value: status},
			},
		}
	}

	// Cross-partition query across all users.
	pager := containerClient.NewQueryItemsPager(query, azcosmos.PartitionKey{}, queryOptions)

	var complaints []models.Complaint
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			s.log.Error("failed to query all complaints", slog.String("status", status), slog.String("error", err.Error()))
			return nil, err
		}

		for _, item := range page.Items {
			var complaint models.Complaint
			if err := json.Unmarshal(item, &complaint); err != nil {
				s.log.Error("failed to unmarshal complaint", slog.String("error", err.Error()))
				return nil, err
			}
			complaints = append(complaints, complaint)
		}
	}

	s.log.Debug("all complaints retrieved", slog.String("status", status), slog.Int("count", len(complaints)))
	return complaints, nil
}

// DeleteComplaint deletes a complaint by ID using the partition key
func (s *Service) DeleteComplaint(ctx context.Context, complaintID string) error {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
		s.log.Error("failed to get complaints container", slog.String("error", err.Error()))
		return err
	}

	// First, find the complaint to get the partition key (userId) and verify it exists
	complaint, err := s.GetComplaintByID(ctx, complaintID)
	if err != nil {
		s.log.Error("failed to get complaint for deletion", slog.String("complaintId", complaintID), slog.String("error", err.Error()))
		return err
	}

	if complaint == nil {
		s.log.Debug("complaint not found for deletion", slog.String("complaintId", complaintID))
		return ErrComplaintNotFound
	}

	// Delete the item using the partition key
	partitionKey := azcosmos.NewPartitionKeyString(complaint.UserID)
	_, err = containerClient.DeleteItem(ctx, partitionKey, complaintID, nil)
	if err != nil {
		s.log.Error("failed to delete complaint from cosmos", slog.String("complaintId", complaintID), slog.String("userId", complaint.UserID), slog.String("error", err.Error()))
		return err
	}

	s.log.Info("complaint deleted successfully", slog.String("complaintId", complaintID), slog.String("userId", complaint.UserID))
	return nil
}
