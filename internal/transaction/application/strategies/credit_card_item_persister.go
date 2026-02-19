package strategies

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

// CreditCardItemPersister define o contrato para persistir itens CREDIT_CARD
// de um MonthlyTransaction após uma sincronização de faturas.
//
// Responsabilidade única: decidir INSERT ou UPDATE para cada item CREDIT_CARD
// com base no snapshot de IDs que existiam antes da mutação do aggregate.
type CreditCardItemPersister interface {
	Persist(
		ctx context.Context,
		tx database.DBTX,
		aggregate *entities.MonthlyTransaction,
		existingIDs map[string]struct{},
	) error
}

type creditCardItemPersister struct {
	repo interfaces.TransactionRepository
}

// NewCreditCardItemPersister cria uma nova instância do persister.
func NewCreditCardItemPersister(repo interfaces.TransactionRepository) CreditCardItemPersister {
	return &creditCardItemPersister{repo: repo}
}

// Persist percorre os itens do aggregate e persiste apenas os do tipo CREDIT_CARD.
//   - ID ausente do snapshot → INSERT (item criado pela mutação do aggregate)
//   - ID presente no snapshot → UPDATE (item existente, inclusive soft-delete)
//   - Itens de outros tipos são ignorados (responsabilidade de outra camada)
func (p *creditCardItemPersister) Persist(
	ctx context.Context,
	tx database.DBTX,
	aggregate *entities.MonthlyTransaction,
	existingIDs map[string]struct{},
) error {
	for _, item := range aggregate.Items {
		if !item.Type.IsCreditCard() {
			continue
		}

		_, alreadyExists := existingIDs[item.ID.String()]

		if !alreadyExists {
			if err := p.repo.InsertItem(ctx, tx, item); err != nil {
				return fmt.Errorf("failed to insert credit card item: %w", err)
			}
			continue
		}

		if err := p.repo.UpdateItem(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to update credit card item: %w", err)
		}
	}

	return nil
}
