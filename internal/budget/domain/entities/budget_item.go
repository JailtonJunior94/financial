package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
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
		AmountUsed:      vos.NewMoney(0),
		PercentageUsed:  vos.NewPercentage(0),
		PercentageTotal: vos.NewPercentage(100),
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}

	budgetItem.CalculateAmountGoal()
	return budgetItem
}

func (b *BudgetItem) CalculateAmountGoal() {
	b.AmountGoal = b.Budget.AmountGoal.Mul(b.PercentageGoal.Percentage())
}

func (b *BudgetItem) AddAmountUsed(amount vos.Money) error {
	b.AmountUsed = b.AmountUsed.Add(amount)

	// Prevent division by zero
	if b.Budget.AmountGoal.Money() == 0 {
		b.PercentageUsed = vos.NewPercentage(0)
		b.PercentageTotal = vos.NewPercentage(0)
		return nil
	}

	// Calculate percentage used based on actual amount used
	percentageUsed, err := b.AmountUsed.Div(b.Budget.AmountGoal.Money())
	if err != nil {
		return err
	}
	b.PercentageUsed = vos.NewPercentage(percentageUsed.Mul(100).Money())

	// Calculate percentage of this item relative to total budget
	percentageTotal, err := b.AmountUsed.Div(b.Budget.AmountGoal.Money())
	if err != nil {
		return err
	}
	b.PercentageTotal = vos.NewPercentage(percentageTotal.Mul(100).Money())

	return nil
}
