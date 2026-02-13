package dtos

import (
	"time"
)

// PaymentMethodPaginationMeta contém os metadados de paginação para métodos de pagamento.
type PaymentMethodPaginationMeta struct {
	Limit      int     `json:"limit"                  example:"20"`
	HasNext    bool    `json:"has_next"               example:"false"`
	NextCursor *string `json:"next_cursor,omitempty"  example:"eyJmIjp7ImNvZGUiOiJQSVgifX0"`
}

// PaymentMethodPaginatedOutput é a resposta paginada de métodos de pagamento (usada na documentação Swagger).
type PaymentMethodPaginatedOutput struct {
	Data       []PaymentMethodOutput       `json:"data"`
	Pagination PaymentMethodPaginationMeta `json:"pagination"`
}

type (
	PaymentMethodInput struct {
		Name        string `json:"name"                example:"Pix"`
		Code        string `json:"code"                example:"PIX"`
		Description string `json:"description,omitempty" example:"Transferência instantânea via Pix"`
	}

	PaymentMethodUpdateInput struct {
		Name        string `json:"name"                example:"Pix Atualizado"`
		Description string `json:"description,omitempty" example:"Pagamento instantâneo 24h"`
	}

	PaymentMethodOutput struct {
		ID          string    `json:"id"                    example:"550e8400-e29b-41d4-a716-446655440000"`
		Name        string    `json:"name"                  example:"Pix"`
		Code        string    `json:"code"                  example:"PIX"`
		Description string    `json:"description,omitempty" example:"Transferência instantânea via Pix"`
		CreatedAt   time.Time `json:"created_at"            example:"2025-01-15T10:30:00Z"`
		UpdatedAt   time.Time `json:"updated_at"            example:"2025-01-20T08:00:00Z"`
	}
)
