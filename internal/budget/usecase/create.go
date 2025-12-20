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
		o11y             o11y.Telemetry
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
	o11y o11y.Telemetry,
	budgetRepository interfaces.BudgetRepository,
) CreateBudgetUseCase {
	return &createBudgetUseCase{
		uow:              uow,
		o11y:             o11y,
		budgetRepository: budgetRepository,
	}
}

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BugetInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_budget_usecase.execute")
	defer span.End()

	newBudget, err := factories.CreateBudget(userID, input)
	if err != nil {
		span.AddEvent("error creating budget entity", o11y.Attribute{Key: "user_id", Value: userID}, o11y.Attribute{Key: "error", Value: err})
		u.o11y.Logger().Error(ctx, err, "error creating budget entity", o11y.Field{Key: "user_id", Value: userID})
		return nil, err
	}

	err = u.uow.Do(ctx, func(ctx context.Context) error {
		if err := u.budgetRepository.Insert(ctx, newBudget); err != nil {
			span.AddEvent("error inserting budget", o11y.Attribute{Key: "user_id", Value: userID}, o11y.Attribute{Key: "error", Value: err})
			u.o11y.Logger().Error(ctx, err, "error inserting budget", o11y.Field{Key: "user_id", Value: userID})
			return err
		}

		if err := u.budgetRepository.InsertItems(ctx, newBudget.Items); err != nil {
			span.AddEvent("error inserting budget items", o11y.Attribute{Key: "user_id", Value: userID}, o11y.Attribute{Key: "error", Value: err})
			u.o11y.Logger().Error(ctx, err, "error inserting budget items", o11y.Field{Key: "user_id", Value: userID})
			return err
		}
		return nil
	})

	if err != nil {
		span.AddEvent("error in unit of work transaction", o11y.Attribute{Key: "user_id", Value: userID}, o11y.Attribute{Key: "error", Value: err})
		u.o11y.Logger().Error(ctx, err, "error in unit of work transaction", o11y.Field{Key: "user_id", Value: userID})
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
