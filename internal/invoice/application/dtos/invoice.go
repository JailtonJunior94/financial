package dtos

import "time"

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

// PurchaseUpdateInput representa o input para atualizar uma compra.
type PurchaseUpdateInput struct {
	CategoryID  string `json:"category_id"`
	Description string `json:"description"`
	TotalAmount string `json:"total_amount"`
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
