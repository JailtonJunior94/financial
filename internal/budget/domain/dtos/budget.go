package dtos

import "time"

type (
	BugetInput struct {
		Date       time.Time     `json:"date"`
		AmountGoal float64       `json:"amount"`
		Items      []*BudgetItem `json:"items"`
	}

	BugetItemInput struct {
		Amount     float64 `json:"amount"`
		CategoryID string  `json:"category_id"`
	}

	BudgetItem struct {
		CategoryID     string  `json:"category_id"`
		PercentageGoal float64 `json:"percentage_goal"`
	}

	BudgetOutput struct {
		ID         string    `json:"id"`
		Date       time.Time `json:"date"`
		AmountGoal float64   `json:"amount_goal"`
		AmountUsed float64   `json:"amount_used"`
		Percentage float64   `json:"percentage"`
		CreatedAt  time.Time `json:"created_at"`
	}
)
