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
	Likes       []string  `json:"likes,omitempty"` // Array of user IDs who liked this complaint
	LikeCount   int       `json:"likeCount"`       // Total number of likes
	CreatedAt   time.Time `json:"createdAt"`
}

// ComplaintResponse is the response DTO for complaints with user-specific like information
type ComplaintResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Comments    []Comment `json:"comments,omitempty"`
	LikeCount   int       `json:"likeCount"` // Total number of likes
	IsLiked     bool      `json:"isLiked"`   // Whether the current user liked this complaint
	CreatedAt   time.Time `json:"createdAt"`
}

const (
	StatusPending  string = "pending"
	StatusApproved string = "approved"
	StatusRejected string = "rejected"
)
