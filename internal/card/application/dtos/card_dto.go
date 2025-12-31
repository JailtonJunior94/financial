package dtos

import (
	"time"
)

type (
	CardInput struct {
		Name   string `json:"name"`
		DueDay int    `json:"due_day"`
	}

	CardOutput struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		DueDay    int       `json:"due_day"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
)
