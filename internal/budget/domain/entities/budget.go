package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type Budget struct {
	entity.Base
	UserID         vos.UUID
	Date           time.Time
	AmountGoal     vos.Money
	AmountUsed     vos.Money
	PercentageUsed vos.Percentage
	Items          []*BudgetItem
}

func NewBudget(userID vos.UUID, amountGoal vos.Money, date time.Time) *Budget {
	return &Budget{
		Date:       date,
		UserID:     userID,
		AmountGoal: amountGoal,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}

func (b *Budget) AddItems(items []*BudgetItem) bool {
	b.Items = append(b.Items, items...)
	b.CalculateAmountUsed()
	b.CalculatePercentageUsed()
	return b.CalculatePercentageTotal()
}

func (b *Budget) AddItem(item *BudgetItem) bool {
	b.Items = append(b.Items, item)
	b.CalculateAmountUsed()
	b.CalculatePercentageUsed()
	return b.CalculatePercentageTotal()
}

func (b *Budget) CalculateAmountUsed() {
	for _, item := range b.Items {
		if sum, err := b.AmountUsed.Add(item.AmountUsed); err == nil {
			b.AmountUsed = sum
		}
	}
}

func (b *Budget) CalculatePercentageUsed() {
	for _, item := range b.Items {
		if sum, err := b.PercentageUsed.Add(item.PercentageUsed); err == nil {
			b.PercentageUsed = sum
		}
	}
}

func (b *Budget) CalculatePercentageTotal() bool {
	var total vos.Percentage
	for _, item := range b.Items {
		if sum, err := total.Add(item.PercentageGoal); err == nil {
			total = sum
		}
	}
	// NewPercentage(100*1000) because scale is 3 decimal places (100.000%)
	hundredPercent, _ := vos.NewPercentage(100000)
	return total.Equals(hundredPercent)
}
