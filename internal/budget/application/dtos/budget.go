package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// BudgetCreateInput representa o input para criar um orçamento.
type BudgetCreateInput struct {
	ReferenceMonth string            `json:"reference_month"` // YYYY-MM format
	TotalAmount    string            `json:"total_amount"`    // String decimal (e.g., "5000.00")
	Currency       string            `json:"currency"`        // ISO 4217 (e.g., "BRL")
	Items          []BudgetItemInput `json:"items"`
}

// Validate valida os campos do input.
func (b *BudgetCreateInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// ReferenceMonth
	if !validation.IsRequired(b.ReferenceMonth) {
		errs.Add("reference_month", "is required")
	}
	if !validation.IsMonth(b.ReferenceMonth) {
		errs.Add("reference_month", "must be in YYYY-MM format")
	}

	// TotalAmount
	if !validation.IsRequired(b.TotalAmount) {
		errs.Add("total_amount", "is required")
	}
	if !validation.IsMoney(b.TotalAmount) {
		errs.Add("total_amount", "must be a valid monetary value")
	}

	// Currency (optional)
	if b.Currency != "" && !validation.IsOneOf(b.Currency, []string{"BRL", "USD", "EUR"}) {
		errs.Add("currency", "must be BRL, USD, or EUR")
	}

	// Items
	if len(b.Items) == 0 {
		errs.Add("items", "at least one item is required")
	} else {
		for i, item := range b.Items {
			itemErrs := item.Validate()
			for _, err := range itemErrs {
				errs.Add(err.Field+"["+string(rune(i))+"]", err.Message)
			}
		}
	}

	return errs
}

// BudgetUpdateInput representa o input para atualizar um orçamento.
type BudgetUpdateInput struct {
	TotalAmount string            `json:"total_amount"` // String decimal
	Items       []BudgetItemInput `json:"items"`
}

// Validate valida os campos do input.
func (b *BudgetUpdateInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// TotalAmount
	if !validation.IsRequired(b.TotalAmount) {
		errs.Add("total_amount", "is required")
	}
	if !validation.IsMoney(b.TotalAmount) {
		errs.Add("total_amount", "must be a valid monetary value")
	}

	// Items
	if len(b.Items) == 0 {
		errs.Add("items", "at least one item is required")
	} else {
		for i, item := range b.Items {
			itemErrs := item.Validate()
			for _, err := range itemErrs {
				errs.Add(err.Field+"["+string(rune(i))+"]", err.Message)
			}
		}
	}

	return errs
}

// BudgetItemInput representa um item de orçamento no input.
type BudgetItemInput struct {
	CategoryID     string `json:"category_id"`
	PercentageGoal string `json:"percentage_goal"` // String decimal (e.g., "25.50")
}

// Validate valida os campos do BudgetItemInput.
func (b *BudgetItemInput) Validate() validation.ValidationErrors {
	var errs validation.ValidationErrors

	// CategoryID
	if !validation.IsRequired(b.CategoryID) {
		errs.Add("category_id", "is required")
	}
	if !validation.IsUUID(b.CategoryID) {
		errs.Add("category_id", "must be a valid UUID")
	}

	// PercentageGoal
	if !validation.IsRequired(b.PercentageGoal) {
		errs.Add("percentage_goal", "is required")
	}
	if !validation.IsMoney(b.PercentageGoal) {
		errs.Add("percentage_goal", "must be a valid percentage value")
	}

	return errs
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
