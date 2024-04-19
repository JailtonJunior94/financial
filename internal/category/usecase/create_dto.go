package usecase

import "time"

type (
	CreateCategoryInput struct {
		Name     string
		Sequence int
	}

	CreateCategoryOutput struct {
		ID        string
		Name      string
		Sequence  int
		CreatedAt time.Time
	}
)
