package strategies

import (
	"errors"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

var (
	ErrCreditCardAmountInvalid    = errors.New("credit card amount must be positive")
	ErrCreditCardMustBeExpense    = errors.New("credit card transactions must be expense")
	ErrCreditCardCannotBeModified = errors.New("credit card items are managed automatically")
)

// CreditCardStrategy implementa validações específicas para faturas de cartão.
//
// REGRA CRÍTICA:
// - Faturas são sempre despesas
// - Devem ser únicas por mês
// - Atualizadas automaticamente quando a fatura muda
type CreditCardStrategy struct{}

// Validate valida uma transação de cartão de crédito.
func (s *CreditCardStrategy) Validate(
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	isPaid bool,
) error {
	if !amount.IsPositive() {
		return ErrCreditCardAmountInvalid
	}

	// Fatura sempre é despesa
	if !direction.IsExpense() {
		return ErrCreditCardMustBeExpense
	}

	return nil
}

// CreateItem cria um novo item de fatura de cartão.
// IMPORTANTE: Normalmente criado via UpdateOrCreateCreditCardItem do aggregate.
func (s *CreditCardStrategy) CreateItem(
	monthlyID sharedVos.UUID,
	categoryID sharedVos.UUID,
	title string,
	description string,
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	isPaid bool,
) (*entities.TransactionItem, error) {
	if err := s.Validate(amount, direction, isPaid); err != nil {
		return nil, err
	}

	return entities.NewTransactionItem(
		monthlyID,
		categoryID,
		title,
		description,
		amount,
		direction,
		transactionVos.TypeCreditCard,
		isPaid,
	)
}
