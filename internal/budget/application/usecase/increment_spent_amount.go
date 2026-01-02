package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// IncrementSpentAmountUseCase incrementa o spent_amount de um budget_item baseado em eventos.
//
// Este use case é usado pelo BudgetUpdateHandler para processar eventos de transação e fatura.
// Diferente do UpdateSpentAmountUseCase (que SUBSTITUI o valor), este INCREMENTA o valor.
//
// Fluxo:
// 1. Busca budget do usuário/mês (retorna erro se não existir)
// 2. Busca budget_item pela categoria
// 3. Incrementa spent_amount (novo valor = atual + increment)
// 4. Recalcula percentageUsed automaticamente
// 5. Persiste as alterações
type (
	IncrementSpentAmountUseCase interface {
		Execute(ctx context.Context, userID vos.UUID, referenceMonth budgetVos.ReferenceMonth, categoryID vos.UUID, incrementAmount vos.Money) error
	}

	incrementSpentAmountUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewIncrementSpentAmountUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) IncrementSpentAmountUseCase {
	return &incrementSpentAmountUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *incrementSpentAmountUseCase) Execute(
	ctx context.Context,
	userID vos.UUID,
	referenceMonth budgetVos.ReferenceMonth,
	categoryID vos.UUID,
	incrementAmount vos.Money,
) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "increment_spent_amount_usecase.execute")
	defer span.End()

	err := u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Criar repositório com a transação
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		// Buscar budget do usuário/mês
		budget, err := budgetRepository.FindByUserIDAndReferenceMonth(ctx, userID, referenceMonth)
		if err != nil {
			return err
		}

		// Se budget não existe, ignora silenciosamente (budget deve ser criado manualmente)
		if budget == nil {
			u.o11y.Logger().Warn(ctx, "budget not found for user/month - ignoring event",
				observability.String("user_id", userID.String()),
				observability.String("reference_month", referenceMonth.String()),
			)
			return nil
		}

		// Buscar budget_item pela categoria
		var targetItem *entities.BudgetItem
		for _, item := range budget.Items {
			if item.CategoryID.String() == categoryID.String() {
				targetItem = item
				break
			}
		}

		// Se não encontrou o item da categoria, ignora silenciosamente
		if targetItem == nil {
			u.o11y.Logger().Warn(ctx, "budget item not found for category - ignoring event",
				observability.String("budget_id", budget.ID.String()),
				observability.String("category_id", categoryID.String()),
			)
			return nil
		}

		// Incrementar spent_amount (novo valor = atual + increment)
		newSpentAmount, err := targetItem.SpentAmount.Add(incrementAmount)
		if err != nil {
			return fmt.Errorf("failed to add increment to spent amount: %w", err)
		}

		// Atualizar através do aggregate root (valida e recalcula totais)
		if err := budget.UpdateItemSpentAmount(targetItem.ID, newSpentAmount); err != nil {
			return err
		}

		// Persistir item atualizado
		if err := budgetRepository.UpdateItem(ctx, targetItem); err != nil {
			return err
		}

		// Persistir totais do budget recalculados
		if err := budgetRepository.Update(ctx, budget); err != nil {
			return err
		}

		u.o11y.Logger().Info(ctx, "budget spent amount incremented",
			observability.String("budget_id", budget.ID.String()),
			observability.String("item_id", targetItem.ID.String()),
			observability.String("category_id", categoryID.String()),
			observability.Int64("increment_cents", incrementAmount.Cents()),
			observability.Int64("new_total_cents", newSpentAmount.Cents()),
		)

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to increment spent amount", observability.Error(err))
		return err
	}

	return nil
}
