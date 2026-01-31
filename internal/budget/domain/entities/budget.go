package entities

import (
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
)

// Constantes de porcentagem usadas em validações e cálculos.
var (
	// hundredPercent representa 100% com scale 3 (100.000).
	// Ignoramos o erro pois sabemos que 100000 é um valor válido.
	hundredPercent, _ = vos.NewPercentage(100000)
	// zeroPercentage representa 0%.
	// Ignoramos o erro pois sabemos que 0 é um valor válido.
	zeroPercentage, _ = vos.NewPercentage(0)
)

const (
	// percentageScale é usado para converter decimal para porcentagem (0.XX * 100 = XX%).
	percentageScale = 100.0
)

// Budget é o Aggregate Root que garante a integridade do orçamento.
type Budget struct {
	entity.Base
	UserID         vos.UUID
	ReferenceMonth budgetVos.ReferenceMonth
	TotalAmount    vos.Money
	SpentAmount    vos.Money
	PercentageUsed vos.Percentage
	Items          []*BudgetItem
}

func NewBudget(userID vos.UUID, totalAmount vos.Money, referenceMonth budgetVos.ReferenceMonth) *Budget {
	zeroMoney, _ := vos.NewMoney(0, totalAmount.Currency())

	return &Budget{
		UserID:         userID,
		ReferenceMonth: referenceMonth,
		TotalAmount:    totalAmount,
		SpentAmount:    zeroMoney,
		PercentageUsed: zeroPercentage,
		Items:          []*BudgetItem{},
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}

// AddItems adiciona múltiplos itens e valida que a soma das porcentagens seja exatamente 100%.
func (b *Budget) AddItems(items []*BudgetItem) error {
	// Valida se items não está vazio
	if len(items) == 0 {
		return domain.ErrBudgetNoItems
	}

	// Calcula soma total das porcentagens incluindo os novos itens
	var totalPercentage vos.Percentage
	for _, existingItem := range b.Items {
		sum, err := totalPercentage.Add(existingItem.PercentageGoal)
		if err != nil {
			return err
		}
		totalPercentage = sum
	}

	for _, newItem := range items {
		// Valida categoria duplicada
		if b.hasCategoryID(newItem.CategoryID) {
			return domain.ErrDuplicateCategory
		}

		sum, err := totalPercentage.Add(newItem.PercentageGoal)
		if err != nil {
			return err
		}
		totalPercentage = sum
	}

	// Valida que soma seja exatamente 100%
	if totalPercentage.GreaterThan(hundredPercent) {
		return domain.ErrBudgetPercentageExceeds100
	}
	if !totalPercentage.Equals(hundredPercent) {
		return domain.ErrBudgetInvalidTotal
	}

	// Adiciona os itens
	b.Items = append(b.Items, items...)
	b.recalculateSpentAmount()
	b.recalculatePercentageUsed()

	return nil
}

// AddItem adiciona um único item e valida que a soma das porcentagens não exceda 100%.
func (b *Budget) AddItem(item *BudgetItem) error {
	// Valida categoria duplicada
	if b.hasCategoryID(item.CategoryID) {
		return domain.ErrDuplicateCategory
	}

	// Calcula soma total das porcentagens incluindo o novo item
	var totalPercentage vos.Percentage
	for _, existingItem := range b.Items {
		sum, err := totalPercentage.Add(existingItem.PercentageGoal)
		if err != nil {
			return err
		}
		totalPercentage = sum
	}

	sum, err := totalPercentage.Add(item.PercentageGoal)
	if err != nil {
		return err
	}
	totalPercentage = sum

	// Valida que soma não exceda 100%
	if totalPercentage.GreaterThan(hundredPercent) {
		return domain.ErrBudgetPercentageExceeds100
	}

	// Adiciona o item
	b.Items = append(b.Items, item)
	b.recalculateSpentAmount()
	b.recalculatePercentageUsed()

	return nil
}

// UpdateItemSpentAmount atualiza o valor gasto de um item específico (passando pelo aggregate).
func (b *Budget) UpdateItemSpentAmount(itemID vos.UUID, newSpentAmount vos.Money) error {
	// Valida que o valor não seja negativo
	if newSpentAmount.IsNegative() {
		return domain.ErrNegativeAmount
	}

	// Encontra o item
	item := b.findItemByID(itemID)
	if item == nil {
		return domain.ErrBudgetItemNotFound
	}

	// PERMITE gastar acima do planejado - o RemainingAmount ficará negativo
	// Isso é comportamento esperado: usuário pode estourar o orçamento

	// Atualiza o valor gasto do item
	item.SpentAmount = newSpentAmount
	item.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

	// Recalcula os totais do budget
	b.recalculateSpentAmount()
	b.recalculatePercentageUsed()

	return nil
}

// FindItemByID busca um item pelo ID.
func (b *Budget) FindItemByID(itemID vos.UUID) *BudgetItem {
	return b.findItemByID(itemID)
}

// hasCategoryID verifica se já existe um item com a categoria informada.
func (b *Budget) hasCategoryID(categoryID vos.UUID) bool {
	for _, item := range b.Items {
		if item.CategoryID.String() == categoryID.String() {
			return true
		}
	}
	return false
}

// findItemByID busca um item pelo ID.
func (b *Budget) findItemByID(itemID vos.UUID) *BudgetItem {
	for _, item := range b.Items {
		if item.ID.String() == itemID.String() {
			return item
		}
	}
	return nil
}

// recalculateSpentAmount recalcula o valor total gasto.
func (b *Budget) recalculateSpentAmount() {
	zeroCurrency := b.TotalAmount.Currency()
	total, _ := vos.NewMoney(0, zeroCurrency)

	for _, item := range b.Items {
		if sum, err := total.Add(item.SpentAmount); err == nil {
			total = sum
		}
	}
	b.SpentAmount = total
}

// recalculatePercentageUsed recalcula a porcentagem total utilizada.
func (b *Budget) recalculatePercentageUsed() {
	// Evita divisão por zero
	if b.TotalAmount.IsZero() {
		b.PercentageUsed = zeroPercentage
		return
	}

	// Calcula: (SpentAmount / TotalAmount) * 100
	spentFloat := b.SpentAmount.Float()
	totalFloat := b.TotalAmount.Float()
	percentageFloat := (spentFloat / totalFloat) * percentageScale

	percentageUsed, err := vos.NewPercentageFromFloat(percentageFloat)
	if err != nil {
		b.PercentageUsed = zeroPercentage
		return
	}

	b.PercentageUsed = percentageUsed
}

// TotalPercentageAllocated retorna a porcentagem total alocada nos itens.
// Se ocorrer erro ao somar (overflow), retorna o total calculado até aquele ponto.
func (b *Budget) TotalPercentageAllocated() vos.Percentage {
	var total vos.Percentage
	for _, item := range b.Items {
		sum, err := total.Add(item.PercentageGoal)
		if err != nil {
			// Erro ao somar - retorna total parcial
			// Indica possível overflow ou dados inválidos
			return total
		}
		total = sum
	}
	return total
}

// IsFullyAllocated verifica se o orçamento está 100% alocado.
func (b *Budget) IsFullyAllocated() bool {
	return b.TotalPercentageAllocated().Equals(hundredPercent)
}

// Delete marca o budget como deletado (soft delete).
func (b *Budget) Delete() *Budget {
	b.DeletedAt = vos.NewNullableTime(time.Now().UTC())
	return b
}
