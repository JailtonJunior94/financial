package dtos

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/validation"
)

// BudgetPaginationMeta contém os metadados de paginação para orçamentos.
type BudgetPaginationMeta struct {
	Limit      int     `json:"limit"                  example:"20"`
	HasNext    bool    `json:"has_next"               example:"false"`
	NextCursor *string `json:"next_cursor,omitempty"  example:"eyJmIjp7InJlZmVyZW5jZV9tb250aCI6IjIwMjUtMDEifX0"`
}

// BudgetPaginatedOutput é a resposta paginada de orçamentos (usada na documentação Swagger).
type BudgetPaginatedOutput struct {
	Data       []BudgetOutput       `json:"data"`
	Pagination BudgetPaginationMeta `json:"pagination"`
}

// BudgetCreateInput representa o input para criar um orçamento.
type BudgetCreateInput struct {
	ReferenceMonth string            `json:"reference_month" example:"2025-01"`                        // YYYY-MM format
	TotalAmount    string            `json:"total_amount"    example:"5000.00"`                         // String decimal (e.g., "5000.00")
	Currency       string            `json:"currency"        example:"BRL" enums:"BRL,USD,EUR"`         // ISO 4217 (e.g., "BRL")
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
	TotalAmount string            `json:"total_amount" example:"6000.00"` // String decimal
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
	CategoryID     string `json:"category_id"     example:"550e8400-e29b-41d4-a716-446655440000"`
	PercentageGoal string `json:"percentage_goal" example:"25.50"` // String decimal (e.g., "25.50")
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
	if !validation.IsPercentage(b.PercentageGoal) {
		errs.Add("percentage_goal", "must be a valid percentage value (up to 3 decimal places)")
	}

	return errs
}

// UpdateSpentAmountInput representa o input para atualizar o valor gasto de um item.
type UpdateSpentAmountInput struct {
	SpentAmount string `json:"spent_amount"` // String decimal
}

// BudgetOutput representa a resposta de um orçamento.
type BudgetOutput struct {
	ID             string             `json:"id"              example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         string             `json:"user_id"         example:"660e8400-e29b-41d4-a716-446655440001"`
	ReferenceMonth string             `json:"reference_month" example:"2025-01"` // YYYY-MM
	TotalAmount    string             `json:"total_amount"    example:"5000.00"`
	SpentAmount    string             `json:"spent_amount"    example:"2350.00"`
	PercentageUsed string             `json:"percentage_used" example:"47.000"`
	Currency       string             `json:"currency"        example:"BRL"      enums:"BRL,USD,EUR"`
	Items          []BudgetItemOutput `json:"items,omitempty"`
	CreatedAt      time.Time          `json:"created_at"      example:"2025-01-01T00:00:00Z"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
}

// BudgetItemOutput representa a resposta de um item de orçamento.
type BudgetItemOutput struct {
	ID              string    `json:"id"               example:"770e8400-e29b-41d4-a716-446655440002"`
	BudgetID        string    `json:"budget_id"        example:"550e8400-e29b-41d4-a716-446655440000"`
	CategoryID      string    `json:"category_id"      example:"880e8400-e29b-41d4-a716-446655440003"`
	PercentageGoal  string    `json:"percentage_goal"  example:"30.000"`
	PlannedAmount   string    `json:"planned_amount"   example:"1500.00"`
	SpentAmount     string    `json:"spent_amount"     example:"700.00"`
	RemainingAmount string    `json:"remaining_amount" example:"800.00"`
	PercentageSpent string    `json:"percentage_spent" example:"46.670"`
	CreatedAt       time.Time `json:"created_at"       example:"2025-01-01T00:00:00Z"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" example:"2025-01-20T08:00:00Z"`
}
