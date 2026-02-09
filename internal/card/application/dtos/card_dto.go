package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

type (
	CardInput struct {
		Name              string `json:"name"`
		DueDay            int    `json:"due_day"`
		ClosingOffsetDays int    `json:"closing_offset_days"` // Opcional, padrão: 7
	}

	CardOutput struct {
		ID                string    `json:"id"`
		Name              string    `json:"name"`
		DueDay            int       `json:"due_day"`
		ClosingOffsetDays int       `json:"closing_offset_days"` // Quantos dias antes do vencimento fecha (padrão brasileiro: 7)
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at,omitempty"`
	}
)

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
