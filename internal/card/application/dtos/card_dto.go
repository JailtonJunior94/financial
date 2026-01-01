package dtos

import (
	"time"
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
