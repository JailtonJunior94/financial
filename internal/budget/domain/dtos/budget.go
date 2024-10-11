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
		CategoryID string  `json:"categoryId"`
	}

	BudgetItem struct {
		CategoryID     string  `json:"categoryId"`
		PercentageGoal float64 `json:"percentageGoal"`
	}

	BudgetOutput struct {
		ID         string    `json:"id"`
		Date       time.Time `json:"date"`
		AmountGoal float64   `json:"amountGoal"`
		AmountUsed float64   `json:"amountUsed"`
		Percentage float64   `json:"percentage"`
		CreatedAt  time.Time `json:"createdAt"`
	}
)
