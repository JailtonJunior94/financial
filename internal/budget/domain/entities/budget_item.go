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
		Budget:          budget,
		BudgetID:        budget.ID,
		CategoryID:      categoryID,
		PercentageGoal:  percentageGoal,
		PercentageUsed:  vos.NewPercentage(0),
		PercentageTotal: vos.NewPercentage(100),
		AmountUsed:      vos.NewMoney(0),
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

func (b *BudgetItem) AddAmountUsed(amount vos.Money) {
	b.AmountUsed = b.AmountUsed.Add(amount)
	b.PercentageUsed = b.PercentageUsed.Add(b.PercentageGoal)

	total, _ := b.AmountUsed.Div(b.Budget.Amount.Money())
	b.PercentageTotal = vos.NewPercentage(total.Mul(100).Money())
}
