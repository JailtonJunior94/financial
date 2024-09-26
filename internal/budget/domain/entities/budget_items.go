package entities

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/entity"
	"github.com/jailtonjunior94/financial/pkg/vos"
)

type BudgetItem struct {
	entity.Base
	Budget            *Budget
	BudgetID          vos.UUID
	CategoryID        vos.UUID
	AmountSpent       vos.Money
	AmountGoal        vos.Money
	PercentageUsedWas vos.Percentage
	PercentageTotal   vos.Percentage
}

func NewBudgetItem(
	budget *Budget,
	budgetID,
	categoryID vos.UUID,
	amountSpent,
	amountGoal vos.Money,
	percentageUsedWas,
	percentageTotal vos.Percentage,
) *BudgetItem {
	return &BudgetItem{
		Budget:            budget,
		BudgetID:          budgetID,
		CategoryID:        categoryID,
		AmountSpent:       amountSpent,
		AmountGoal:        amountGoal,
		PercentageUsedWas: percentageUsedWas,
		PercentageTotal:   percentageTotal,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}
