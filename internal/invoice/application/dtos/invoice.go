package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// PurchaseCreateInput representa o input para criar uma compra.
type PurchaseCreateInput struct {
	CardID           string `json:"card_id"`
	CategoryID       string `json:"category_id"`
	PurchaseDate     string `json:"purchase_date"` // YYYY-MM-DD
	Description      string `json:"description"`
	TotalAmount      string `json:"total_amount"`      // String decimal (e.g., "1200.00")
	Currency         string `json:"currency"`          // ISO 4217 (e.g., "BRL")
	InstallmentTotal int    `json:"installment_total"` // 1 para à vista
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
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	TotalAmount string `json:"total_amount"`
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
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	CardID         string              `json:"card_id"`
	ReferenceMonth string              `json:"reference_month"` // YYYY-MM
	DueDate        string              `json:"due_date"`        // YYYY-MM-DD
	TotalAmount    string              `json:"total_amount"`
	Currency       string              `json:"currency"`
	ItemCount      int                 `json:"item_count"`
	Items          []InvoiceItemOutput `json:"items,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at,omitempty"`
}

// InvoiceItemOutput representa a resposta de um item de fatura.
type InvoiceItemOutput struct {
	ID                string    `json:"id"`
	InvoiceID         string    `json:"invoice_id"`
	CategoryID        string    `json:"category_id"`
	PurchaseDate      string    `json:"purchase_date"` // YYYY-MM-DD
	Description       string    `json:"description"`
	TotalAmount       string    `json:"total_amount"`       // Valor total da compra original
	InstallmentNumber int       `json:"installment_number"` // 1 a N
	InstallmentTotal  int       `json:"installment_total"`  // Total de parcelas
	InstallmentAmount string    `json:"installment_amount"` // Valor desta parcela
	InstallmentLabel  string    `json:"installment_label"`  // Ex: "3/12" ou "À vista"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// InvoiceListOutput representa uma lista resumida de faturas.
type InvoiceListOutput struct {
	ID             string    `json:"id"`
	CardID         string    `json:"card_id"`
	ReferenceMonth string    `json:"reference_month"`
	DueDate        string    `json:"due_date"`
	TotalAmount    string    `json:"total_amount"`
	Currency       string    `json:"currency"`
	ItemCount      int       `json:"item_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// PurchaseCreateOutput representa a resposta ao criar uma compra com parcelas.
type PurchaseCreateOutput struct {
	Items []InvoiceItemOutput `json:"items"`
}

// PurchaseUpdateOutput representa a resposta ao atualizar uma compra.
type PurchaseUpdateOutput struct {
	Items []InvoiceItemOutput `json:"items"`
}
