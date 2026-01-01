package strategies

import (
	"errors"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

var (
	ErrPixAmountInvalid = errors.New("PIX amount must be positive")
)

// PixStrategy implementa validações específicas para transações PIX.
type PixStrategy struct{}

// Validate valida uma transação PIX.
func (s *PixStrategy) Validate(
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	isPaid bool,
) error {
	if !amount.IsPositive() {
		return ErrPixAmountInvalid
	}
	return nil
}

// CreateItem cria um novo item de transação PIX.
func (s *PixStrategy) CreateItem(
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
		transactionVos.TypePix,
		isPaid,
	)
}
