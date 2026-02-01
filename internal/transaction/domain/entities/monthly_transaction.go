package entities

import (
	"slices"
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// MonthlyTransaction é o Aggregate Root que representa o consolidado financeiro mensal.
//
// Responsabilidades:
// - Manter todas as invariantes.
// - Centralizar todos os cálculos.
// - Garantir consistência dos totais.
// - Gerenciar completamente os TransactionItems.
type MonthlyTransaction struct {
	ID             sharedVos.UUID
	UserID         sharedVos.UUID
	ReferenceMonth transactionVos.ReferenceMonth
	TotalIncome    sharedVos.Money
	TotalExpense   sharedVos.Money
	TotalAmount    sharedVos.Money
	Items          []*TransactionItem
	CreatedAt      sharedVos.NullableTime
	UpdatedAt      sharedVos.NullableTime
}

// NewMonthlyTransaction cria um novo MonthlyTransaction.
func NewMonthlyTransaction(
	userID sharedVos.UUID,
	referenceMonth transactionVos.ReferenceMonth,
) (*MonthlyTransaction, error) {
	// Cria o aggregate com valores zerados
	currency := sharedVos.CurrencyBRL
	zero, _ := sharedVos.NewMoney(0, currency)

	return &MonthlyTransaction{
		UserID:         userID,
		ReferenceMonth: referenceMonth,
		TotalIncome:    zero,
		TotalExpense:   zero,
		TotalAmount:    zero,
		Items:          make([]*TransactionItem, 0),
		CreatedAt:      sharedVos.NewNullableTime(time.Now()),
	}, nil
}

// AddItem adiciona um novo item e recalcula os totais.
// REGRA CRÍTICA: Items CREDIT_CARD devem ser únicos por mês.
func (m *MonthlyTransaction) AddItem(item *TransactionItem) error {
	if item == nil {
		return ErrItemNotFound
	}

	// Validação: CREDIT_CARD deve ser único
	if item.Type.IsCreditCard() && m.hasCreditCardItem() {
		return ErrCreditCardItemAlreadyExists
	}

	m.Items = append(m.Items, item)
	m.recalculateTotals()
	m.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

// UpdateItem atualiza um item existente e recalcula os totais.
// O aggregate garante que o item pertence a ele.
func (m *MonthlyTransaction) UpdateItem(
	itemID sharedVos.UUID,
	title string,
	description string,
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	transactionType transactionVos.TransactionType,
	isPaid bool,
) error {
	item := m.findItemByID(itemID)
	if item == nil {
		return ErrItemNotFound
	}

	if item.MonthlyID.String() != m.ID.String() {
		return ErrItemDoesNotBelong
	}

	if err := item.Update(title, description, amount, direction, transactionType, isPaid); err != nil {
		return err
	}

	m.recalculateTotals()
	m.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

// RemoveItem remove um item (soft delete) e recalcula os totais.
func (m *MonthlyTransaction) RemoveItem(itemID sharedVos.UUID) error {
	item := m.findItemByID(itemID)
	if item == nil {
		return ErrItemNotFound
	}

	if item.MonthlyID.String() != m.ID.String() {
		return ErrItemDoesNotBelong
	}

	item.Delete()
	m.recalculateTotals()
	m.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

// UpdateOrCreateCreditCardItem atualiza ou cria o item de cartão de crédito.
// Garante idempotência: apenas um item CREDIT_CARD por mês.
func (m *MonthlyTransaction) UpdateOrCreateCreditCardItem(
	categoryID sharedVos.UUID,
	amount sharedVos.Money,
	isPaid bool,
) error {
	// Procura item CREDIT_CARD existente
	existingItem := m.findCreditCardItem()

	if existingItem != nil {
		// Atualiza o item existente
		return m.UpdateItem(
			existingItem.ID,
			"Fatura do Cartão de Crédito",
			"Consolidado mensal da fatura",
			amount,
			transactionVos.DirectionExpense,
			transactionVos.TypeCreditCard,
			isPaid,
		)
	}

	// Cria novo item
	newItem, err := NewTransactionItem(
		m.ID,
		categoryID,
		"Fatura do Cartão de Crédito",
		"Consolidado mensal da fatura",
		amount,
		transactionVos.DirectionExpense,
		transactionVos.TypeCreditCard,
		isPaid,
	)
	if err != nil {
		return err
	}

	// Gera ID para o novo item
	id, _ := sharedVos.NewUUID()
	newItem.SetID(id)

	return m.AddItem(newItem)
}

// recalculateTotals recalcula todos os totais do aggregate.
// REGRA CRÍTICA: Sempre executado após add/update/remove.
func (m *MonthlyTransaction) recalculateTotals() {
	currency := sharedVos.CurrencyBRL
	zero, _ := sharedVos.NewMoney(0, currency)

	income := zero
	expense := zero

	// Itera apenas sobre itens não deletados
	for _, item := range m.activeItems() {
		if item.Direction.IsIncome() {
			income, _ = income.Add(item.Amount)
		} else if item.Direction.IsExpense() {
			expense, _ = expense.Add(item.Amount)
		}
	}

	m.TotalIncome = income
	m.TotalExpense = expense
	m.TotalAmount, _ = income.Subtract(expense)
}

// activeItems retorna apenas itens não deletados.
func (m *MonthlyTransaction) activeItems() []*TransactionItem {
	active := make([]*TransactionItem, 0)
	for _, item := range m.Items {
		if !item.IsDeleted() {
			active = append(active, item)
		}
	}
	return active
}

// findItemByID encontra um item pelo ID.
func (m *MonthlyTransaction) findItemByID(itemID sharedVos.UUID) *TransactionItem {
	idx := slices.IndexFunc(m.Items, func(item *TransactionItem) bool {
		return item.ID.String() == itemID.String()
	})
	if idx == -1 {
		return nil
	}
	return m.Items[idx]
}

// findCreditCardItem encontra o item CREDIT_CARD (se existir).
func (m *MonthlyTransaction) findCreditCardItem() *TransactionItem {
	idx := slices.IndexFunc(m.activeItems(), func(item *TransactionItem) bool {
		return item.Type.IsCreditCard()
	})
	if idx == -1 {
		return nil
	}
	return m.activeItems()[idx]
}

// hasCreditCardItem verifica se já existe um item CREDIT_CARD ativo.
func (m *MonthlyTransaction) hasCreditCardItem() bool {
	return m.findCreditCardItem() != nil
}

// SetID define o ID do aggregate (usado apenas pelo repositório).
func (m *MonthlyTransaction) SetID(id sharedVos.UUID) {
	m.ID = id
}

// LoadItems carrega os items no aggregate (usado apenas pelo repositório).
func (m *MonthlyTransaction) LoadItems(items []*TransactionItem) {
	m.Items = items
	m.recalculateTotals()
}
