package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/factories"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/database"
	"github.com/jailtonjunior94/financial/pkg/database/uow"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.BugetInput) (*dtos.BudgetOutput, error)
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

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BugetInput) (*dtos.BudgetOutput, error) {
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
	return &dtos.BudgetOutput{
		ID:         newBudget.ID.String(),
		Date:       newBudget.Date,
		AmountGoal: newBudget.AmountGoal.Float(),
		AmountUsed: newBudget.AmountUsed.Float(),
		Percentage: newBudget.PercentageUsed.Float(),
		CreatedAt:  newBudget.CreatedAt,
	}, nil
}
