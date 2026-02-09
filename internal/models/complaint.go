package models

import "time"

type Complaint struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

const (
	StatusPending  string = "pending"
	StatusApproved string = "approved"
	StatusRejected string = "rejected"
)
