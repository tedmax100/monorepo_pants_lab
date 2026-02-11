// Package models provides shared data models for all Go services.
package models

import "time"

// User represents a user in the system.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Demo      string    `json:"demo,omitempty"`
}

// Order represents an order in the system.
type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Product   string    `json:"product"`
	Quantity  int       `json:"quantity"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}
