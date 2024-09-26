package entities

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/entity"
	"github.com/jailtonjunior94/financial/pkg/vos"
)

type BudgetItem struct {
	entity.Base
	Budget          *Budget
	BudgetID        vos.UUID
	CategoryID      vos.UUID
	PercentageGoal  vos.Percentage
	AmountGoal      vos.Money
	AmountUsed      vos.Money
	PercentageUsed  vos.Percentage
	PercentageTotal vos.Percentage
}

func NewBudgetItem(
	budget *Budget,
	categoryID vos.UUID,
	percentageGoal vos.Percentage,
) *BudgetItem {
	budgetItem := &BudgetItem{
		Budget:         budget,
		BudgetID:       budget.ID,
		CategoryID:     categoryID,
		PercentageGoal: percentageGoal,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}

	budgetItem.CalculateAmountGoal()
	return budgetItem
}

func (b *BudgetItem) CalculateAmountGoal() {
	b.AmountGoal = b.Budget.Amount.Mul(b.PercentageGoal.Percentage())
}
