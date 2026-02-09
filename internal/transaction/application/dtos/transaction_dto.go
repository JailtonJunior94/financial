package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// RegisterTransactionInput representa os dados para registrar uma nova transação.
type RegisterTransactionInput struct {
	ReferenceMonth string `json:"reference_month"` // Formato: YYYY-MM
	CategoryID     string `json:"category_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Amount         string `json:"amount"`    // String decimal (e.g., "1234.56")
	Direction      string `json:"direction"` // INCOME | EXPENSE
	Type           string `json:"type"`      // PIX | BOLETO | TRANSFER | CREDIT_CARD
	IsPaid         bool   `json:"is_paid"`
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
	Title       string `json:"title"`
	Description string `json:"description"`
	Amount      string `json:"amount"` // String decimal (e.g., "1234.56")
	Direction   string `json:"direction"`
	Type        string `json:"type"`
	IsPaid      bool   `json:"is_paid"`
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
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Amount      string    `json:"amount"` // String decimal (e.g., "1234.56")
	Direction   string    `json:"direction"`
	Type        string    `json:"type"`
	IsPaid      bool      `json:"is_paid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// MonthlyTransactionOutput representa o consolidado mensal na resposta.
type MonthlyTransactionOutput struct {
	ID             string                   `json:"id"`
	ReferenceMonth string                   `json:"reference_month"`
	TotalIncome    string                   `json:"total_income"`  // String decimal (e.g., "5000.00")
	TotalExpense   string                   `json:"total_expense"` // String decimal (e.g., "3500.00")
	TotalAmount    string                   `json:"total_amount"`  // String decimal (e.g., "1500.00")
	Items          []*TransactionItemOutput `json:"items"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at,omitempty"`
}
