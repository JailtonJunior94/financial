package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// BudgetItem representa um item individual de um orçamento.
// Nota: Mutações devem passar pelo Budget (aggregate root).
type BudgetItem struct {
	entity.Base
	BudgetID       vos.UUID
	CategoryID     vos.UUID
	PercentageGoal vos.Percentage
	PlannedAmount  vos.Money
	SpentAmount    vos.Money
}

func NewBudgetItem(
	budgetID vos.UUID,
	budgetTotalAmount vos.Money,
	categoryID vos.UUID,
	percentageGoal vos.Percentage,
) *BudgetItem {
	// Initialize zero values with correct types
	zeroCurrency := budgetTotalAmount.Currency()
	zeroMoney, _ := vos.NewMoney(0, zeroCurrency)

	// Calculate planned amount
	plannedAmount, _ := percentageGoal.Apply(budgetTotalAmount)

	return &BudgetItem{
		BudgetID:       budgetID,
		CategoryID:     categoryID,
		PercentageGoal: percentageGoal,
		PlannedAmount:  plannedAmount,
		SpentAmount:    zeroMoney,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}

// PercentageSpent calcula a porcentagem gasta em relação ao planejado.
func (b *BudgetItem) PercentageSpent() vos.Percentage {
	// Evita divisão por zero
	if b.PlannedAmount.IsZero() {
		zeroPercentage, _ := vos.NewPercentage(0)
		return zeroPercentage
	}

	// Calcula: (SpentAmount / PlannedAmount) * 100
	spentFloat := b.SpentAmount.Float()
	plannedFloat := b.PlannedAmount.Float()
	percentageFloat := (spentFloat / plannedFloat) * percentageScale

	percentageSpent, err := vos.NewPercentageFromFloat(percentageFloat)
	if err != nil {
		zeroPercentage, _ := vos.NewPercentage(0)
		return zeroPercentage
	}

	return percentageSpent
}

// RemainingAmount calcula o valor restante disponível.
func (b *BudgetItem) RemainingAmount() vos.Money {
	remaining, err := b.PlannedAmount.Subtract(b.SpentAmount)
	if err != nil {
		zeroCurrency := b.PlannedAmount.Currency()
		zeroMoney, _ := vos.NewMoney(0, zeroCurrency)
		return zeroMoney
	}
	return remaining
}
