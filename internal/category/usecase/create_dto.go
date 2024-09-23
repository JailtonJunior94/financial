package usecase

import (
	"time"
)

type (
	CreateCategoryInput struct {
		ParentID string
		Name     string
		Sequence uint
	}

	CreateCategoryOutput struct {
		ID        string
		Name      string
		Sequence  uint
		CreatedAt time.Time
	}
)
