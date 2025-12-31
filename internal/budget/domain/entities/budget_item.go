package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// BudgetItem representa um item individual de um orçamento
// Nota: Mutações devem passar pelo Budget (aggregate root)
type BudgetItem struct {
	entity.Base
	Budget         *Budget
	BudgetID       vos.UUID
	CategoryID     vos.UUID
	PercentageGoal vos.Percentage
	PlannedAmount  vos.Money
	SpentAmount    vos.Money
}

func NewBudgetItem(
	budget *Budget,
	categoryID vos.UUID,
	percentageGoal vos.Percentage,
) *BudgetItem {
	// Initialize zero values with correct types
	zeroCurrency := budget.TotalAmount.Currency()
	zeroMoney, _ := vos.NewMoney(0, zeroCurrency)

	budgetItem := &BudgetItem{
		Budget:         budget,
		BudgetID:       budget.ID,
		CategoryID:     categoryID,
		PercentageGoal: percentageGoal,
		SpentAmount:    zeroMoney,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}

	budgetItem.calculatePlannedAmount()
	return budgetItem
}

// calculatePlannedAmount calcula o valor planejado com base na porcentagem
func (b *BudgetItem) calculatePlannedAmount() {
	// Apply percentage to budget total amount
	if amount, err := b.PercentageGoal.Apply(b.Budget.TotalAmount); err == nil {
		b.PlannedAmount = amount
	}
}

// PercentageSpent calcula a porcentagem gasta em relação ao planejado
func (b *BudgetItem) PercentageSpent() vos.Percentage {
	// Evita divisão por zero
	if b.PlannedAmount.IsZero() {
		zeroPercentage, _ := vos.NewPercentage(0)
		return zeroPercentage
	}

	// Calcula: (SpentAmount / PlannedAmount) * 100
	spentFloat := b.SpentAmount.Float()
	plannedFloat := b.PlannedAmount.Float()
	percentageFloat := (spentFloat / plannedFloat) * 100.0

	percentageSpent, err := vos.NewPercentageFromFloat(percentageFloat)
	if err != nil {
		zeroPercentage, _ := vos.NewPercentage(0)
		return zeroPercentage
	}

	return percentageSpent
}

// RemainingAmount calcula o valor restante disponível
func (b *BudgetItem) RemainingAmount() vos.Money {
	remaining, err := b.PlannedAmount.Subtract(b.SpentAmount)
	if err != nil {
		zeroCurrency := b.PlannedAmount.Currency()
		zeroMoney, _ := vos.NewMoney(0, zeroCurrency)
		return zeroMoney
	}
	return remaining
}
