package dtos

import (
	"regexp"
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

var lastFourDigitsPattern = regexp.MustCompile(`^[0-9]{4}$`)

type (
	CardInput struct {
		Name              string `json:"name"                          example:"Nubank Platinum"`
		Type              string `json:"type"                          example:"credit"`
		Flag              string `json:"flag"                          example:"mastercard"`
		LastFourDigits    string `json:"last_four_digits"              example:"7890"`
		DueDay            *int   `json:"due_day,omitempty"             example:"10"`
		ClosingOffsetDays *int   `json:"closing_offset_days,omitempty" example:"7"`
	}

	CardUpdateInput struct {
		Name              string `json:"name"                          example:"Nubank Platinum"`
		Flag              string `json:"flag"                          example:"mastercard"`
		LastFourDigits    string `json:"last_four_digits"              example:"7890"`
		DueDay            *int   `json:"due_day,omitempty"             example:"10"`
		ClosingOffsetDays *int   `json:"closing_offset_days,omitempty" example:"7"`
	}

	CardOutput struct {
		ID                string    `json:"id"                            example:"550e8400-e29b-41d4-a716-446655440000"`
		Name              string    `json:"name"                          example:"Nubank Platinum"`
		Type              string    `json:"type"                          example:"credit"`
		Flag              string    `json:"flag"                          example:"mastercard"`
		LastFourDigits    string    `json:"last_four_digits"              example:"7890"`
		DueDay            *int      `json:"due_day,omitempty"             example:"10"`
		ClosingOffsetDays *int      `json:"closing_offset_days,omitempty" example:"7"`
		CreatedAt         time.Time `json:"created_at"                    example:"2025-01-15T10:30:00Z"`
		UpdatedAt         time.Time `json:"updated_at,omitempty"          example:"2025-01-20T08:00:00Z"`
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

	if !validation.IsRequired(c.Name) {
		errs.Add("name", "is required")
	}
	if !validation.IsMaxLength(c.Name, 255) {
		errs.Add("name", "must be at most 255 characters")
	}

	if !validation.IsRequired(c.Type) {
		errs.Add("type", "is required")
	} else if !validation.IsOneOf(c.Type, []string{"credit", "debit"}) {
		errs.Add("type", "must be 'credit' or 'debit'")
	}

	if !validation.IsRequired(c.Flag) {
		errs.Add("flag", "is required")
	} else if !validation.IsOneOf(c.Flag, []string{"visa", "mastercard", "elo", "amex", "hipercard"}) {
		errs.Add("flag", "must be one of: visa, mastercard, elo, amex, hipercard")
	}

	if !validation.IsRequired(c.LastFourDigits) {
		errs.Add("last_four_digits", "is required")
	} else if !lastFourDigitsPattern.MatchString(c.LastFourDigits) {
		errs.Add("last_four_digits", "must be exactly 4 numeric digits")
	}

	if c.Type == "credit" {
		if c.DueDay == nil {
			errs.Add("due_day", "is required for credit cards")
		} else if !validation.IsInRange(*c.DueDay, 1, 31) {
			errs.Add("due_day", "must be between 1 and 31")
		}
		if c.ClosingOffsetDays != nil && !validation.IsInRange(*c.ClosingOffsetDays, 1, 31) {
			errs.Add("closing_offset_days", "must be between 1 and 31")
		}
	}

	return errs
}

// Validate valida os campos do CardUpdateInput.
func (c *CardUpdateInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	if !validation.IsRequired(c.Name) {
		errs.Add("name", "is required")
	}
	if !validation.IsMaxLength(c.Name, 255) {
		errs.Add("name", "must be at most 255 characters")
	}

	if !validation.IsRequired(c.Flag) {
		errs.Add("flag", "is required")
	} else if !validation.IsOneOf(c.Flag, []string{"visa", "mastercard", "elo", "amex", "hipercard"}) {
		errs.Add("flag", "must be one of: visa, mastercard, elo, amex, hipercard")
	}

	if !validation.IsRequired(c.LastFourDigits) {
		errs.Add("last_four_digits", "is required")
	} else if !lastFourDigitsPattern.MatchString(c.LastFourDigits) {
		errs.Add("last_four_digits", "must be exactly 4 numeric digits")
	}

	if c.DueDay != nil && !validation.IsInRange(*c.DueDay, 1, 31) {
		errs.Add("due_day", "must be between 1 and 31")
	}
	if c.ClosingOffsetDays != nil && !validation.IsInRange(*c.ClosingOffsetDays, 1, 31) {
		errs.Add("closing_offset_days", "must be between 1 and 31")
	}

	return errs
}
