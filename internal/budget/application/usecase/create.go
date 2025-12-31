package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/factories"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/database"
	"github.com/jailtonjunior94/financial/pkg/database/uow"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.BudgetCreateInput) (*dtos.BudgetOutput, error)
	}

	createBudgetUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewCreateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) CreateBudgetUseCase {
	return &createBudgetUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BudgetCreateInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_budget_usecase.execute")
	defer span.End()

	newBudget, err := factories.CreateBudget(userID, input)
	if err != nil {
		return nil, err
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBExecutor) error {
		// Criar repositório com a transação
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		if err := budgetRepository.Insert(ctx, newBudget); err != nil {
			return err
		}

		if err := budgetRepository.InsertItems(ctx, newBudget.Items); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Build output
	return &dtos.BudgetOutput{
		ID:             newBudget.ID.String(),
		UserID:         newBudget.UserID.String(),
		ReferenceMonth: newBudget.ReferenceMonth.String(),
		TotalAmount:    newBudget.TotalAmount.String(),
		SpentAmount:    newBudget.SpentAmount.String(),
		PercentageUsed: newBudget.PercentageUsed.String(),
		Currency:       string(newBudget.TotalAmount.Currency()),
		CreatedAt:      newBudget.CreatedAt,
	}, nil
}
