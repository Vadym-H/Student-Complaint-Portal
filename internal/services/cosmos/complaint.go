package cosmos

import (
	"context"
	"encoding/json"

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
			return nil, err
		}

		for _, item := range page.Items {
			var complaint models.Complaint
			if err := json.Unmarshal(item, &complaint); err != nil {
				return nil, err
			}
			complaints = append(complaints, complaint)
		}
	}

	return complaints, nil
}

// UpdateComplaintStatus updates the status of a complaint by ID
func (s *Service) UpdateComplaintStatus(ctx context.Context, id, status string) error {
	containerClient, err := s.client.NewContainer(s.database, s.complaintsContainer)
	if err != nil {
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
			return err
		}

		for _, item := range page.Items {
			var c models.Complaint
			if err := json.Unmarshal(item, &c); err != nil {
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
		return nil // complaint not found
	}

	// Update the status
	complaint.Status = status

	// Marshal the updated complaint
	complaintBytes, err := json.Marshal(complaint)
	if err != nil {
		return err
	}

	// Replace the item using the partition key
	partitionKey := azcosmos.NewPartitionKeyString(complaint.UserID)
	_, err = containerClient.ReplaceItem(ctx, partitionKey, id, complaintBytes, nil)
	return err
}
