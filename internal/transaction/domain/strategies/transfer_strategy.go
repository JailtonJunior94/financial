package strategies

import (
	"errors"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

var (
	ErrTransferAmountInvalid = errors.New("transfer amount must be positive")
)

// TransferStrategy implementa validações específicas para transferências.
type TransferStrategy struct{}

// Validate valida uma transferência.
func (s *TransferStrategy) Validate(
	amount sharedVos.Money,
	direction transactionVos.TransactionDirection,
	isPaid bool,
) error {
	if !amount.IsPositive() {
		return ErrTransferAmountInvalid
	}
	return nil
}

// CreateItem cria um novo item de transferência.
func (s *TransferStrategy) CreateItem(
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
		transactionVos.TypeTransfer,
		isPaid,
	)
}
