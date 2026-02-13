package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// PurchaseCreateInput representa o input para criar uma compra.
type PurchaseCreateInput struct {
	CardID           string `json:"card_id"           example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID       string `json:"category_id"       example:"660e8400-e29b-41d4-a716-446655440001"`
	PurchaseDate     string `json:"purchase_date"     example:"2025-01-15"`                              // YYYY-MM-DD
	Description      string `json:"description"       example:"iPhone 16 Pro"`
	TotalAmount      string `json:"total_amount"      example:"9999.00"`                                  // String decimal (e.g., "1200.00")
	Currency         string `json:"currency"          example:"BRL" enums:"BRL,USD,EUR"`                  // ISO 4217 (e.g., "BRL")
	InstallmentTotal int    `json:"installment_total" example:"12"  minimum:"1" maximum:"48"`             // 1 para à vista
}

// Validate valida os campos do input.
func (i *PurchaseCreateInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// CardID
	if !validation.IsRequired(i.CardID) {
		errs.Add("card_id", "is required")
	}

	if !validation.IsUUID(i.CardID) {
		errs.Add("card_id", "must be a valid UUID")
	}

	// CategoryID
	if !validation.IsRequired(i.CategoryID) {
		errs.Add("category_id", "is required")
	}
	if !validation.IsUUID(i.CategoryID) {
		errs.Add("category_id", "must be a valid UUID")
	}

	// PurchaseDate
	if !validation.IsRequired(i.PurchaseDate) {
		errs.Add("purchase_date", "is required")
	}
	if !validation.IsDate(i.PurchaseDate) {
		errs.Add("purchase_date", "must be in YYYY-MM-DD format")
	}

	// Description
	if !validation.IsRequired(i.Description) {
		errs.Add("description", "is required")
	}
	if !validation.IsMaxLength(i.Description, 255) {
		errs.Add("description", "must be at most 255 characters")
	}

	// TotalAmount
	if !validation.IsRequired(i.TotalAmount) {
		errs.Add("total_amount", "is required")
	}
	if !validation.IsMoney(i.TotalAmount) {
		errs.Add("total_amount", "must be a valid monetary value")
	}

	// Currency (optional, defaults to BRL)
	if i.Currency != "" && !validation.IsOneOf(i.Currency, []string{"BRL", "USD", "EUR"}) {
		errs.Add("currency", "must be BRL, USD, or EUR")
	}

	// InstallmentTotal
	if !validation.IsPositiveInt(i.InstallmentTotal) {
		errs.Add("installment_total", "must be at least 1")
	}
	if !validation.IsInRange(i.InstallmentTotal, 1, 48) {
		errs.Add("installment_total", "must be between 1 and 48")
	}

	return errs
}

// PurchaseUpdateInput representa o input para atualizar uma compra.
type PurchaseUpdateInput struct {
	CategoryID  string `json:"category_id"  example:"660e8400-e29b-41d4-a716-446655440001"`
	Description string `json:"description"  example:"iPhone 16 Pro Max"`
	TotalAmount string `json:"total_amount" example:"10999.00"`
}

// Validate valida os campos do input.
func (i *PurchaseUpdateInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// CategoryID
	if !validation.IsRequired(i.CategoryID) {
		errs.Add("category_id", "is required")
	} else if !validation.IsUUID(i.CategoryID) {
		errs.Add("category_id", "must be a valid UUID")
	}

	// Description
	if !validation.IsRequired(i.Description) {
		errs.Add("description", "is required")
	} else if !validation.IsMaxLength(i.Description, 255) {
		errs.Add("description", "must be at most 255 characters")
	}

	// TotalAmount
	if !validation.IsRequired(i.TotalAmount) {
		errs.Add("total_amount", "is required")
	} else if !validation.IsMoney(i.TotalAmount) {
		errs.Add("total_amount", "must be a valid monetary value")
	}

	return errs
}

// InvoiceOutput representa a resposta de uma fatura.
type InvoiceOutput struct {
	ID             string              `json:"id"              example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         string              `json:"user_id"         example:"660e8400-e29b-41d4-a716-446655440001"`
	CardID         string              `json:"card_id"         example:"770e8400-e29b-41d4-a716-446655440002"`
	ReferenceMonth string              `json:"reference_month" example:"2025-01"` // YYYY-MM
	DueDate        string              `json:"due_date"        example:"2025-01-10"` // YYYY-MM-DD
	TotalAmount    string              `json:"total_amount"    example:"9999.00"`
	Currency       string              `json:"currency"        example:"BRL" enums:"BRL,USD,EUR"`
	ItemCount      int                 `json:"item_count"      example:"12"`
	Items          []InvoiceItemOutput `json:"items,omitempty"`
	CreatedAt      time.Time           `json:"created_at"      example:"2025-01-01T00:00:00Z"`
	UpdatedAt      time.Time           `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
}

// InvoiceItemOutput representa a resposta de um item de fatura.
type InvoiceItemOutput struct {
	ID                string    `json:"id"                 example:"880e8400-e29b-41d4-a716-446655440003"`
	InvoiceID         string    `json:"invoice_id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID        string    `json:"category_id"        example:"660e8400-e29b-41d4-a716-446655440001"`
	PurchaseDate      string    `json:"purchase_date"      example:"2025-01-15"` // YYYY-MM-DD
	Description       string    `json:"description"        example:"iPhone 16 Pro"`
	TotalAmount       string    `json:"total_amount"       example:"9999.00"` // Valor total da compra original
	InstallmentNumber int       `json:"installment_number" example:"3"` // 1 a N
	InstallmentTotal  int       `json:"installment_total"  example:"12"` // Total de parcelas
	InstallmentAmount string    `json:"installment_amount" example:"833.25"` // Valor desta parcela
	InstallmentLabel  string    `json:"installment_label"  example:"3/12"` // Ex: "3/12" ou "À vista"
	CreatedAt         time.Time `json:"created_at"         example:"2025-01-01T00:00:00Z"`
	UpdatedAt         time.Time `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
}

// InvoiceListOutput representa uma lista resumida de faturas.
type InvoiceListOutput struct {
	ID             string    `json:"id"              example:"550e8400-e29b-41d4-a716-446655440000"`
	CardID         string    `json:"card_id"         example:"770e8400-e29b-41d4-a716-446655440002"`
	ReferenceMonth string    `json:"reference_month" example:"2025-01"`
	DueDate        string    `json:"due_date"        example:"2025-01-10"`
	TotalAmount    string    `json:"total_amount"    example:"9999.00"`
	Currency       string    `json:"currency"        example:"BRL" enums:"BRL,USD,EUR"`
	ItemCount      int       `json:"item_count"      example:"12"`
	CreatedAt      time.Time `json:"created_at"      example:"2025-01-01T00:00:00Z"`
}

// PurchaseCreateOutput representa a resposta ao criar uma compra com parcelas.
type PurchaseCreateOutput struct {
	Items []InvoiceItemOutput `json:"items"`
}

// PurchaseUpdateOutput representa a resposta ao atualizar uma compra.
type PurchaseUpdateOutput struct {
	Items []InvoiceItemOutput `json:"items"`
}

// InvoicePaginationMeta contém os metadados de paginação para faturas.
type InvoicePaginationMeta struct {
	Limit      int     `json:"limit"                  example:"20"`
	HasNext    bool    `json:"has_next"               example:"true"`
	NextCursor *string `json:"next_cursor,omitempty"  example:"eyJmIjp7InJlZmVyZW5jZV9tb250aCI6IjIwMjUtMDEifX0"`
}

// InvoicePaginatedOutput é a resposta paginada de faturas (usada na documentação Swagger).
type InvoicePaginatedOutput struct {
	Data       []InvoiceListOutput   `json:"data"`
	Pagination InvoicePaginationMeta `json:"pagination"`
}
