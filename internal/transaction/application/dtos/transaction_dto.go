package dtos

import "time"

// RegisterTransactionInput representa os dados para registrar uma nova transação.
type RegisterTransactionInput struct {
	ReferenceMonth string  `json:"reference_month"` // Formato: YYYY-MM
	CategoryID     string  `json:"category_id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Direction      string  `json:"direction"` // INCOME | EXPENSE
	Type           string  `json:"type"`      // PIX | BOLETO | TRANSFER | CREDIT_CARD
	IsPaid         bool    `json:"is_paid"`
}

// UpdateTransactionItemInput representa os dados para atualizar um item.
type UpdateTransactionItemInput struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Direction   string  `json:"direction"`
	Type        string  `json:"type"`
	IsPaid      bool    `json:"is_paid"`
}

// TransactionItemOutput representa um item de transação na resposta.
type TransactionItemOutput struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
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
	TotalIncome    float64                  `json:"total_income"`
	TotalExpense   float64                  `json:"total_expense"`
	TotalAmount    float64                  `json:"total_amount"`
	Items          []*TransactionItemOutput `json:"items"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at,omitempty"`
}
