package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	CreateBudgetItemUseCase interface {
		Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetItemInput) error
	}

	createBudgetItemUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewCreateBudgetItemUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) CreateBudgetItemUseCase {
	return &createBudgetItemUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *createBudgetItemUseCase) Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetItemInput) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_budget_item_usecase.execute")
	defer span.End()

	// Parse userID
	uid, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid user_id: %w", err)
	}

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid budget_id: %w", err)
	}

	// Parse category ID
	categoryID, err := vos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid category_id: %w", err)
	}

	// Parse percentage goal
	percentageGoal, err := vos.NewPercentageFromString(input.PercentageGoal)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid percentage_goal: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		// 1. Find the budget by ID (scoped by userID to prevent IDOR)
		budget, err := budgetRepository.FindByID(ctx, uid, id)
		if err != nil {
			span.RecordError(err)
			return err
		}

		if budget == nil {
			span.RecordError(domain.ErrBudgetNotFound)
			return domain.ErrBudgetNotFound
		}

		// 2. Create a new BudgetItem entity
		newItem := entities.NewBudgetItem(budget.ID, budget.TotalAmount, categoryID, percentageGoal)

		// 3. Call budget.AddItem() to validate and add
		if err := budget.AddItem(newItem); err != nil {
			span.RecordError(err)
			return err
		}

		// 4. Persist the item
		if err := budgetRepository.InsertItems(ctx, []*entities.BudgetItem{newItem}); err != nil {
			span.RecordError(err)
			return err
		}

		// Update budget totals after adding item
		if err := budgetRepository.Update(ctx, budget); err != nil {
			span.RecordError(err)
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
