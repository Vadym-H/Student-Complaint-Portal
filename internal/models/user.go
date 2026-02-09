package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"passwordHash,omitempty"`
	Role         string    `json:"role"` // "student" or "admin"
	CreatedAt    time.Time `json:"createdAt"`
}

const (
	RoleStudent string = "student"
	RoleAdmin   string = "admin"
)
