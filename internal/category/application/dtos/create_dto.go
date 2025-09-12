package dtos

import (
	"time"
)

type (
	CategoryInput struct {
		ParentID string `json:"parent_id"`
		Name     string `json:"name"`
		Sequence uint   `json:"sequence"`
	}

	CategoryOutput struct {
		ID        string           `json:"id"`
		Name      string           `json:"name"`
		Sequence  uint             `json:"sequence"`
		CreatedAt time.Time        `json:"created_at"`
		Children  []CategoryOutput `json:"children,omitempty"`
	}
)
