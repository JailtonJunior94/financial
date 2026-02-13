package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

type (
	CardInput struct {
		Name              string `json:"name"                example:"Nubank Platinum"`
		DueDay            int    `json:"due_day"             example:"10" minimum:"1" maximum:"31"`
		ClosingOffsetDays int    `json:"closing_offset_days" example:"7"  minimum:"1" maximum:"31"` // Opcional, padrão: 7
	}

	CardOutput struct {
		ID                string    `json:"id"                  example:"550e8400-e29b-41d4-a716-446655440000"`
		Name              string    `json:"name"                example:"Nubank Platinum"`
		DueDay            int       `json:"due_day"             example:"10"`
		ClosingOffsetDays int       `json:"closing_offset_days" example:"7"` // Quantos dias antes do vencimento fecha (padrão brasileiro: 7)
		CreatedAt         time.Time `json:"created_at"          example:"2025-01-15T10:30:00Z"`
		UpdatedAt         time.Time `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
	}
)

// CardPaginationMeta contém os metadados de paginação para cards.
type CardPaginationMeta struct {
	Limit      int     `json:"limit"                  example:"20"`
	HasNext    bool    `json:"has_next"               example:"true"`
	NextCursor *string `json:"next_cursor,omitempty"  example:"eyJmIjp7Im5hbWUiOiJOdWJhbmsifX0"`
}

// CardPaginatedOutput é a resposta paginada de cartões (usada na documentação Swagger).
type CardPaginatedOutput struct {
	Data       []CardOutput       `json:"data"`
	Pagination CardPaginationMeta `json:"pagination"`
}

// Validate valida os campos do CardInput.
func (c *CardInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// Name
	if !validation.IsRequired(c.Name) {
		errs.Add("name", "is required")
	}
	if !validation.IsMaxLength(c.Name, 255) {
		errs.Add("name", "must be at most 255 characters")
	}

	// DueDay
	if !validation.IsPositiveInt(c.DueDay) {
		errs.Add("due_day", "must be at least 1")
	}
	if !validation.IsInRange(c.DueDay, 1, 31) {
		errs.Add("due_day", "must be between 1 and 31")
	}

	// ClosingOffsetDays (optional, 0 means use default)
	if c.ClosingOffsetDays != 0 {
		if !validation.IsPositiveInt(c.ClosingOffsetDays) {
			errs.Add("closing_offset_days", "must be at least 1")
		}
		if !validation.IsInRange(c.ClosingOffsetDays, 1, 31) {
			errs.Add("closing_offset_days", "must be between 1 and 31")
		}
	}

	return errs
}
