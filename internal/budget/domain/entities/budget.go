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
		b.AmountUsed = b.AmountUsed.Add(item.AmountUsed)
	}
}

func (b *Budget) CalculatePercentageUsed() {
	for _, item := range b.Items {
		b.PercentageUsed = b.PercentageUsed.Add(item.PercentageUsed)
	}
}

func (b *Budget) CalculatePercentageTotal() bool {
	var total vos.Percentage
	for _, item := range b.Items {
		total = total.Add(item.PercentageGoal)
	}
	return total.Equals(vos.NewPercentage(100))
}
