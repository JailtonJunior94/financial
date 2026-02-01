package dtos

import "time"

// BudgetCreateInput representa o input para criar um orçamento.
type BudgetCreateInput struct {
	ReferenceMonth string            `json:"reference_month"` // YYYY-MM format
	TotalAmount    string            `json:"total_amount"`    // String decimal (e.g., "5000.00")
	Currency       string            `json:"currency"`        // ISO 4217 (e.g., "BRL")
	Items          []BudgetItemInput `json:"items"`
}

// BudgetUpdateInput representa o input para atualizar um orçamento.
type BudgetUpdateInput struct {
	TotalAmount string            `json:"total_amount"` // String decimal
	Items       []BudgetItemInput `json:"items"`
}

// BudgetItemInput representa um item de orçamento no input.
type BudgetItemInput struct {
	CategoryID     string `json:"category_id"`
	PercentageGoal string `json:"percentage_goal"` // String decimal (e.g., "25.50")
}

// UpdateSpentAmountInput representa o input para atualizar o valor gasto de um item.
type UpdateSpentAmountInput struct {
	SpentAmount string `json:"spent_amount"` // String decimal
}

// BudgetOutput representa a resposta de um orçamento.
type BudgetOutput struct {
	ID             string             `json:"id"`
	UserID         string             `json:"user_id"`
	ReferenceMonth string             `json:"reference_month"` // YYYY-MM
	TotalAmount    string             `json:"total_amount"`
	SpentAmount    string             `json:"spent_amount"`
	PercentageUsed string             `json:"percentage_used"`
	Currency       string             `json:"currency"`
	Items          []BudgetItemOutput `json:"items,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty"`
}

// BudgetItemOutput representa a resposta de um item de orçamento.
type BudgetItemOutput struct {
	ID              string    `json:"id"`
	BudgetID        string    `json:"budget_id"`
	CategoryID      string    `json:"category_id"`
	PercentageGoal  string    `json:"percentage_goal"`
	PlannedAmount   string    `json:"planned_amount"`
	SpentAmount     string    `json:"spent_amount"`
	RemainingAmount string    `json:"remaining_amount"`
	PercentageSpent string    `json:"percentage_spent"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}
