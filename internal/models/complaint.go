package models

import "time"

type Comment struct {
	ID        string    `json:"id"`
	AdminID   string    `json:"adminId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type Complaint struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Comments    []Comment `json:"comments,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

const (
	StatusPending  string = "pending"
	StatusApproved string = "approved"
	StatusRejected string = "rejected"
)
