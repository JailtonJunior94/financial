package usecase

import (
	"context"
	"errors"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/factories"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/database/uow"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	CreateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.BugetInput) (*dtos.BudgetOutput, error)
	}

	createBudgetUseCase struct {
		uow              uow.UnitOfWork
		o11y             o11y.Observability
		budgetRepository interfaces.BudgetRepository
	}
)

const (
	BudgetRepository = "BudgetRepository"
)

var (
	ErrInvalidRepositoryType = errors.New("invalid repository type")
)

func NewCreateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y o11y.Observability,
	budgetRepository interfaces.BudgetRepository,
) CreateBudgetUseCase {
	return &createBudgetUseCase{
		uow:              uow,
		o11y:             o11y,
		budgetRepository: budgetRepository,
	}
}

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BugetInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_budget_usecase.execute")
	defer span.End()

	newBudget, err := factories.CreateBudget(userID, input)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	err = u.uow.Do(ctx, func(ctx context.Context) error {
		if err := u.budgetRepository.Insert(ctx, newBudget); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert order", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if err := u.budgetRepository.InsertItems(ctx, newBudget.Items); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert items", o11y.Attributes{Key: "error", Value: err})
			return err
		}
		return nil
	})

	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error insert order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}
	return &dtos.BudgetOutput{
		ID:         newBudget.ID.String(),
		Date:       newBudget.Date,
		AmountGoal: newBudget.AmountGoal.Money(),
		AmountUsed: newBudget.AmountUsed.Money(),
		Percentage: newBudget.PercentageUsed.Percentage(),
		CreatedAt:  newBudget.CreatedAt,
	}, nil
}
