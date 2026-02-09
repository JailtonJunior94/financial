package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
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

// Validate valida os campos do CategoryInput.
func (c *CategoryInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// Name
	if !validation.IsRequired(c.Name) {
		errs.Add("name", "is required")
	} else if !validation.IsMaxLength(c.Name, 255) {
		errs.Add("name", "must be at most 255 characters")
	}

	// ParentID (optional)
	if c.ParentID != "" && !validation.IsUUID(c.ParentID) {
		errs.Add("parent_id", "must be a valid UUID")
	}

	// Sequence
	if c.Sequence == 0 {
		errs.Add("sequence", "must be at least 1")
	}

	return errs
}
