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

// strategyRegistry mapeia cada TransactionType à sua TransactionStrategy correspondente.
// Inicializado com os tipos do domínio; extensível via Register sem alterar este arquivo (OCP).
var strategyRegistry = map[transactionVos.TransactionType]TransactionStrategy{
	transactionVos.TypePix:        &PixStrategy{},
	transactionVos.TypeBoleto:     &BoletoStrategy{},
	transactionVos.TypeTransfer:   &TransferStrategy{},
	transactionVos.TypeCreditCard: &CreditCardStrategy{},
}

// GetStrategy retorna a strategy associada ao tipo informado.
// Retorna nil se o tipo não estiver registrado.
func GetStrategy(transactionType transactionVos.TransactionType) TransactionStrategy {
	return strategyRegistry[transactionType]
}

// Register associa uma TransactionStrategy a um TransactionType.
// Permite adicionar novos tipos de transação sem modificar este arquivo (OCP).
func Register(transactionType transactionVos.TransactionType, strategy TransactionStrategy) {
	strategyRegistry[transactionType] = strategy
}
