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
	// Initialize zero values with correct types
	zeroCurrency := budget.AmountGoal.Currency()
	zeroMoney, _ := vos.NewMoney(0, zeroCurrency)
	zeroPercentage, _ := vos.NewPercentage(0)
	hundredPercentage, _ := vos.NewPercentage(100000) // 100.000% with scale 3

	budgetItem := &BudgetItem{
		Budget:          budget,
		BudgetID:        budget.ID,
		CategoryID:      categoryID,
		PercentageGoal:  percentageGoal,
		AmountUsed:      zeroMoney,
		PercentageUsed:  zeroPercentage,
		PercentageTotal: hundredPercentage,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}

	budgetItem.CalculateAmountGoal()
	return budgetItem
}

func (b *BudgetItem) CalculateAmountGoal() {
	// Apply percentage to budget amount goal
	if amount, err := b.PercentageGoal.Apply(b.Budget.AmountGoal); err == nil {
		b.AmountGoal = amount
	}
}

func (b *BudgetItem) AddAmountUsed(amount vos.Money) error {
	// Add amount used
	newAmountUsed, err := b.AmountUsed.Add(amount)
	if err != nil {
		return err
	}
	b.AmountUsed = newAmountUsed

	// Prevent division by zero
	if b.Budget.AmountGoal.IsZero() {
		zeroPercentage, _ := vos.NewPercentage(0)
		b.PercentageUsed = zeroPercentage
		b.PercentageTotal = zeroPercentage
		return nil
	}

	// Calculate percentage used: (AmountUsed / AmountGoal) * 100
	// Use Float() for division, then convert to percentage with scale 3
	usedFloat := b.AmountUsed.Float()
	goalFloat := b.Budget.AmountGoal.Float()
	percentageFloat := (usedFloat / goalFloat) * 100.0

	// Convert to Percentage (scale 3: 12.345% = 12345)
	percentageUsed, err := vos.NewPercentageFromFloat(percentageFloat)
	if err != nil {
		return err
	}
	b.PercentageUsed = percentageUsed

	// PercentageTotal is the same as PercentageUsed for this item
	b.PercentageTotal = percentageUsed

	return nil
}
