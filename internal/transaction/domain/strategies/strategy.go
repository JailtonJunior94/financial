package strategies

import (
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// TransactionStrategy define o contrato para validação e criação de transaction items.
// Cada tipo de transação implementa suas próprias regras de negócio.
type TransactionStrategy interface {
	// Validate valida os dados da transação.
	Validate(
		amount sharedVos.Money,
		direction transactionVos.TransactionDirection,
		isPaid bool,
	) error

	// CreateItem cria um novo TransactionItem validado.
	CreateItem(
		monthlyID sharedVos.UUID,
		categoryID sharedVos.UUID,
		title string,
		description string,
		amount sharedVos.Money,
		direction transactionVos.TransactionDirection,
		isPaid bool,
	) (*entities.TransactionItem, error)
}

// GetStrategy retorna a strategy apropriada para o tipo de transação.
func GetStrategy(transactionType transactionVos.TransactionType) TransactionStrategy {
	switch transactionType {
	case transactionVos.TypePix:
		return &PixStrategy{}
	case transactionVos.TypeBoleto:
		return &BoletoStrategy{}
	case transactionVos.TypeTransfer:
		return &TransferStrategy{}
	case transactionVos.TypeCreditCard:
		return &CreditCardStrategy{}
	default:
		return nil
	}
}
