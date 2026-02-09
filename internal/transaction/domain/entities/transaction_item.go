package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// TransactionItem representa uma movimentação financeira individual.
// Sempre pertence a um MonthlyTransaction (aggregate root).
type TransactionItem struct {
	ID          sharedVos.UUID
	MonthlyID   sharedVos.UUID // Referência ao aggregate root
	CategoryID  sharedVos.UUID
	Title       string
	Description string
	Amount      sharedVos.Money
	Direction   transactionVos.TransactionDirection
	Type        transactionVos.TransactionType
	IsPaid      bool
	CreatedAt   sharedVos.NullableTime
	UpdatedAt   sharedVos.NullableTime
	DeletedAt   sharedVos.NullableTime
}

// validateTransactionItemFields valida os campos básicos de um TransactionItem
func validateTransactionItemFields(title string, amount sharedVos.Money, direction transactionVos.TransactionDirection, transactionType transactionVos.TransactionType) error {
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > 255 {
		return ErrTitleTooLong
	}
	if !amount.IsPositive() {
		return ErrAmountMustBePositive
	}
	if !direction.IsValid() {
		return ErrInvalidDirection
	}
	if !transactionType.IsValid() {
		return ErrInvalidType
	}
	return nil
}

// NewTransactionItem cria um novo TransactionItem.
// Não deve ser chamado diretamente - use o método do aggregate.
func NewTransactionItem(
	monthlyID sharedVos.UUID,
	categoryID sharedVos.UUID,
	title string,
	description string,
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	transactionType transactionVos.TransactionType,
	isPaid bool,
) (*TransactionItem, error) {
	if err := validateTransactionItemFields(title, amount, direction, transactionType); err != nil {
		return nil, err
	}

	return &TransactionItem{
		MonthlyID:   monthlyID,
		CategoryID:  categoryID,
		Title:       title,
		Description: description,
		Amount:      amount,
		Direction:   direction,
		Type:        transactionType,
		IsPaid:      isPaid,
		CreatedAt:   sharedVos.NewNullableTime(time.Now()),
	}, nil
}

// Update atualiza os dados do item.
// Deve ser chamado através do aggregate para garantir recálculo dos totais.
func (t *TransactionItem) Update(
	title string,
	description string,
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	transactionType transactionVos.TransactionType,
	isPaid bool,
) error {
	if err := validateTransactionItemFields(title, amount, direction, transactionType); err != nil {
		return err
	}

	t.Title = title
	t.Description = description
	t.Amount = amount
	t.Direction = direction
	t.Type = transactionType
	t.IsPaid = isPaid
	t.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

// Delete marca o item como deletado (soft delete).
func (t *TransactionItem) Delete() {
	t.DeletedAt = sharedVos.NewNullableTime(time.Now())
}

// IsDeleted verifica se o item está deletado.
func (t *TransactionItem) IsDeleted() bool {
	return !t.DeletedAt.ValueOr(time.Time{}).IsZero()
}

// SetID define o ID do item (usado apenas pelo repositório).
func (t *TransactionItem) SetID(id sharedVos.UUID) {
	t.ID = id
}
