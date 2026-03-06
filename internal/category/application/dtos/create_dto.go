package dtos

import (
"time"

"github.com/jailtonjunior94/financial/pkg/validation"
)

type (
CategoryInput struct {
Name     string `json:"name"     example:"Alimentacao"`
Sequence uint   `json:"sequence" example:"1" minimum:"1"`
}

CategoryOutput struct {
ID            string              `json:"id"                  example:"550e8400-e29b-41d4-a716-446655440000"`
Name          string              `json:"name"                example:"Alimentacao"`
Sequence      uint                `json:"sequence"            example:"1"`
CreatedAt     time.Time           `json:"created_at"          example:"2025-01-15T10:30:00Z"`
Subcategories []SubcategoryOutput `json:"subcategories,omitempty"`
}
)

type CategoryPaginationMeta struct {
Limit      int     `json:"limit"                  example:"20"`
HasNext    bool    `json:"has_next"               example:"false"`
NextCursor *string `json:"next_cursor,omitempty"  example:"eyJm..."`
}

type CategoryPaginatedOutput struct {
Data       []CategoryOutput       `json:"data"`
Pagination CategoryPaginationMeta `json:"pagination"`
}

func (c *CategoryInput) Validate() validation.ValidationErrors {
var errs validation.ValidationErrors

if !validation.IsRequired(c.Name) {
errs.Add("name", "is required")
} else if !validation.IsMaxLength(c.Name, 255) {
errs.Add("name", "must be at most 255 characters")
}

if c.Sequence == 0 {
errs.Add("sequence", "must be at least 1")
}

return errs
}
