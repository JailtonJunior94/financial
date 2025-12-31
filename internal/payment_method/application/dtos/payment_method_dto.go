package dtos

import (
	"time"
)

type (
	PaymentMethodInput struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description,omitempty"`
	}

	PaymentMethodUpdateInput struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}

	PaymentMethodOutput struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Code        string    `json:"code"`
		Description string    `json:"description,omitempty"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}
)
