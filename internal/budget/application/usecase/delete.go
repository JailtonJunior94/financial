package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	DeleteBudgetUseCase interface {
		Execute(ctx context.Context, budgetID string) error
	}

	deleteBudgetUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewDeleteBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) DeleteBudgetUseCase {
	return &deleteBudgetUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *deleteBudgetUseCase) Execute(ctx context.Context, budgetID string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_budget_usecase.execute")
	defer span.End()

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return fmt.Errorf("invalid budget_id: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		// Find budget
		budget, err := budgetRepository.FindByID(ctx, id)
		if err != nil {
			return err
		}

		if budget == nil {
			return domain.ErrBudgetNotFound
		}

		// Soft delete via repositório (persiste deleted_at no banco)
		if err := budgetRepository.Delete(ctx, id); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
