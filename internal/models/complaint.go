package models

import "time"

type Complaint struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "approved", "rejected"
	CreatedAt   time.Time `json:"createdAt"`
}
