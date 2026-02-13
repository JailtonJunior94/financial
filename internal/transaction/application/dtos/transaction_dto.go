package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// TransactionPaginationMeta contém os metadados de paginação para transações.
type TransactionPaginationMeta struct {
	Limit      int     `json:"limit"                  example:"20"`
	HasNext    bool    `json:"has_next"               example:"true"`
	NextCursor *string `json:"next_cursor,omitempty"  example:"eyJmIjp7InJlZmVyZW5jZV9tb250aCI6IjIwMjUtMDEifX0"`
}

// MonthlyTransactionPaginatedOutput é a resposta paginada de transações mensais (usada na documentação Swagger).
type MonthlyTransactionPaginatedOutput struct {
	Data       []MonthlyTransactionOutput `json:"data"`
	Pagination TransactionPaginationMeta  `json:"pagination"`
}

// RegisterTransactionInput representa os dados para registrar uma nova transação.
type RegisterTransactionInput struct {
	ReferenceMonth string `json:"reference_month" example:"2025-01"`                                    // Formato: YYYY-MM
	CategoryID     string `json:"category_id"     example:"550e8400-e29b-41d4-a716-446655440000"`
	Title          string `json:"title"           example:"Supermercado Pão de Açúcar"`
	Description    string `json:"description"     example:"Compras da semana"`
	Amount         string `json:"amount"          example:"350.75"`                                     // String decimal (e.g., "1234.56")
	Direction      string `json:"direction"       example:"EXPENSE"    enums:"INCOME,EXPENSE"`         // INCOME | EXPENSE
	Type           string `json:"type"            example:"PIX"        enums:"PIX,BOLETO,TRANSFER,CREDIT_CARD"` // PIX | BOLETO | TRANSFER | CREDIT_CARD
	IsPaid         bool   `json:"is_paid"         example:"true"`
}

// Validate valida os campos do input.
func (i *RegisterTransactionInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// ReferenceMonth
	if !validation.IsRequired(i.ReferenceMonth) {
		errs.Add("reference_month", "is required")
	}
	if !validation.IsMonth(i.ReferenceMonth) {
		errs.Add("reference_month", "must be in YYYY-MM format")
	}

	// CategoryID
	if !validation.IsRequired(i.CategoryID) {
		errs.Add("category_id", "is required")
	}
	if validation.IsRequired(i.CategoryID) && !validation.IsUUID(i.CategoryID) {
		errs.Add("category_id", "must be a valid UUID")
	}

	// Title
	if !validation.IsRequired(i.Title) {
		errs.Add("title", "is required")
	}
	if !validation.IsMaxLength(i.Title, 255) {
		errs.Add("title", "must be at most 255 characters")
	}

	// Description (optional)
	if !validation.IsMaxLength(i.Description, 1000) {
		errs.Add("description", "must be at most 1000 characters")
	}

	// Amount
	if !validation.IsRequired(i.Amount) {
		errs.Add("amount", "is required")
	}
	if !validation.IsMoney(i.Amount) {
		errs.Add("amount", "must be a valid monetary value (e.g., 1234.56)")
	}

	// Direction
	if !validation.IsRequired(i.Direction) {
		errs.Add("direction", "is required")
	}
	if !validation.IsOneOf(i.Direction, []string{"INCOME", "EXPENSE"}) {
		errs.Add("direction", "must be INCOME or EXPENSE")
	}

	// Type
	if !validation.IsRequired(i.Type) {
		errs.Add("type", "is required")
	}
	if !validation.IsOneOf(i.Type, []string{"PIX", "BOLETO", "TRANSFER", "CREDIT_CARD"}) {
		errs.Add("type", "must be PIX, BOLETO, TRANSFER, or CREDIT_CARD")
	}

	return errs
}

// UpdateTransactionItemInput representa os dados para atualizar um item.
type UpdateTransactionItemInput struct {
	Title       string `json:"title"       example:"Supermercado Extra"`
	Description string `json:"description" example:"Compras do mês"`
	Amount      string `json:"amount"      example:"420.00"`                                         // String decimal (e.g., "1234.56")
	Direction   string `json:"direction"   example:"EXPENSE" enums:"INCOME,EXPENSE"`
	Type        string `json:"type"        example:"PIX"     enums:"PIX,BOLETO,TRANSFER,CREDIT_CARD"`
	IsPaid      bool   `json:"is_paid"     example:"false"`
}

// Validate valida os campos do input.
func (i *UpdateTransactionItemInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// Title
	if !validation.IsRequired(i.Title) {
		errs.Add("title", "is required")
	} else if !validation.IsMaxLength(i.Title, 255) {
		errs.Add("title", "must be at most 255 characters")
	}

	// Description (optional)
	if !validation.IsMaxLength(i.Description, 1000) {
		errs.Add("description", "must be at most 1000 characters")
	}

	// Amount
	if !validation.IsRequired(i.Amount) {
		errs.Add("amount", "is required")
	} else if !validation.IsMoney(i.Amount) {
		errs.Add("amount", "must be a valid monetary value")
	}

	// Direction
	if !validation.IsRequired(i.Direction) {
		errs.Add("direction", "is required")
	} else if !validation.IsOneOf(i.Direction, []string{"INCOME", "EXPENSE"}) {
		errs.Add("direction", "must be INCOME or EXPENSE")
	}

	// Type
	if !validation.IsRequired(i.Type) {
		errs.Add("type", "is required")
	} else if !validation.IsOneOf(i.Type, []string{"PIX", "BOLETO", "TRANSFER", "CREDIT_CARD"}) {
		errs.Add("type", "must be PIX, BOLETO, TRANSFER, or CREDIT_CARD")
	}

	return errs
}

// TransactionItemOutput representa um item de transação na resposta.
type TransactionItemOutput struct {
	ID          string    `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID  string    `json:"category_id"  example:"660e8400-e29b-41d4-a716-446655440001"`
	Title       string    `json:"title"        example:"Supermercado Pão de Açúcar"`
	Description string    `json:"description"  example:"Compras da semana"`
	Amount      string    `json:"amount"       example:"350.75"` // String decimal (e.g., "1234.56")
	Direction   string    `json:"direction"    example:"EXPENSE" enums:"INCOME,EXPENSE"`
	Type        string    `json:"type"         example:"PIX"     enums:"PIX,BOLETO,TRANSFER,CREDIT_CARD"`
	IsPaid      bool      `json:"is_paid"      example:"true"`
	CreatedAt   time.Time `json:"created_at"   example:"2025-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
}

// MonthlyTransactionOutput representa o consolidado mensal na resposta.
type MonthlyTransactionOutput struct {
	ID             string                   `json:"id"              example:"770e8400-e29b-41d4-a716-446655440002"`
	ReferenceMonth string                   `json:"reference_month" example:"2025-01"`
	TotalIncome    string                   `json:"total_income"    example:"5000.00"` // String decimal (e.g., "5000.00")
	TotalExpense   string                   `json:"total_expense"   example:"3500.00"` // String decimal (e.g., "3500.00")
	TotalAmount    string                   `json:"total_amount"    example:"1500.00"` // String decimal (e.g., "1500.00")
	Items          []*TransactionItemOutput `json:"items"`
	CreatedAt      time.Time                `json:"created_at"      example:"2025-01-01T00:00:00Z"`
	UpdatedAt      time.Time                `json:"updated_at,omitempty" example:"2025-01-31T23:59:59Z"`
}
