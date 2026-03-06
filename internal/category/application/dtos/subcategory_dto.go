package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

type SubcategoryInput struct {
	Name     string `json:"name"     example:"Uber"`
	Sequence uint   `json:"sequence" example:"1" minimum:"1"`
}

type SubcategoryOutput struct {
	ID         string    `json:"id"`
	CategoryID string    `json:"category_id"`
	Name       string    `json:"name"`
	Sequence   uint      `json:"sequence"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubcategoryPaginatedOutput struct {
	Data       []SubcategoryOutput    `json:"data"`
	Pagination CategoryPaginationMeta `json:"pagination"`
}

func (s *SubcategoryInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors
	if !validation.IsRequired(s.Name) {
		errs.Add("name", "is required")
	} else if !validation.IsMaxLength(s.Name, 255) {
		errs.Add("name", "must be at most 255 characters")
	}
	if s.Sequence == 0 {
		errs.Add("sequence", "must be at least 1")
	}
	return errs
}
