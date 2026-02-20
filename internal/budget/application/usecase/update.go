package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/money"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetUpdateInput) (*dtos.BudgetOutput, error)
	}

	updateBudgetUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewUpdateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) UpdateBudgetUseCase {
	return &updateBudgetUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *updateBudgetUseCase) Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetUpdateInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_budget_usecase.execute")
	defer span.End()

	// Parse userID
	uid, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return nil, fmt.Errorf("invalid budget_id: %w", err)
	}

	var updatedBudget *dtos.BudgetOutput

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		// Find budget (scoped by userID to prevent IDOR)
		budget, err := budgetRepository.FindByID(ctx, uid, id)
		if err != nil {
			return err
		}

		if budget == nil {
			return domain.ErrBudgetNotFound
		}

		// Parse new total amount
		newTotalAmount, err := money.NewMoney(input.TotalAmount, budget.TotalAmount.Currency())
		if err != nil {
			return fmt.Errorf("invalid total_amount: %w", err)
		}

		// Validate that total amount is positive
		if newTotalAmount.IsNegative() || newTotalAmount.IsZero() {
			return fmt.Errorf("total_amount must be positive")
		}

		// Update budget total amount
		budget.TotalAmount = newTotalAmount
		budget.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

		// Recalculate planned amounts for all items (since total changed)
		for _, item := range budget.Items {
			// Recalculate PlannedAmount = TotalAmount * PercentageGoal
			plannedAmount, err := item.PercentageGoal.Apply(budget.TotalAmount)
			if err != nil {
				return fmt.Errorf("failed to recalculate planned amount: %w", err)
			}

			item.PlannedAmount = plannedAmount
			item.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

			// Update item in database
			if err := budgetRepository.UpdateItem(ctx, item); err != nil {
				return err
			}
		}

		// Recalculate budget totals (spent amount and percentage used)
		budget.RecalculateTotals()

		// Update budget in database
		if err := budgetRepository.Update(ctx, budget); err != nil {
			return err
		}

		// Build items output
		items := make([]dtos.BudgetItemOutput, len(budget.Items))
		for i, item := range budget.Items {
			items[i] = dtos.BudgetItemOutput{
				ID:              item.ID.String(),
				BudgetID:        item.BudgetID.String(),
				CategoryID:      item.CategoryID.String(),
				PercentageGoal:  fmt.Sprintf("%.3f", item.PercentageGoal.Float()),
				PlannedAmount:   fmt.Sprintf("%.2f", item.PlannedAmount.Float()),
				SpentAmount:     fmt.Sprintf("%.2f", item.SpentAmount.Float()),
				RemainingAmount: fmt.Sprintf("%.2f", item.RemainingAmount().Float()),
				PercentageSpent: fmt.Sprintf("%.3f", item.PercentageSpent().Float()),
				CreatedAt:       item.CreatedAt,
				UpdatedAt:       item.UpdatedAt.ValueOr(time.Time{}),
			}
		}

		// Build output
		updatedBudget = &dtos.BudgetOutput{
			ID:             budget.ID.String(),
			UserID:         budget.UserID.String(),
			ReferenceMonth: budget.ReferenceMonth.String(),
			TotalAmount:    fmt.Sprintf("%.2f", budget.TotalAmount.Float()),
			SpentAmount:    fmt.Sprintf("%.2f", budget.SpentAmount.Float()),
			PercentageUsed: fmt.Sprintf("%.3f", budget.PercentageUsed.Float()),
			Currency:       string(budget.TotalAmount.Currency()),
			Items:          items,
			CreatedAt:      budget.CreatedAt,
			UpdatedAt:      budget.UpdatedAt.ValueOr(time.Time{}),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedBudget, nil
}
