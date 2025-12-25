package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type budgetRepository struct {
	exec database.DBExecutor
	o11y o11y.Telemetry
}

func NewBudgetRepository(exec database.DBExecutor, o11y o11y.Telemetry) interfaces.BudgetRepository {
	return &budgetRepository{
		exec: exec,
		o11y: o11y,
	}
}

func (r *budgetRepository) Insert(ctx context.Context, budget *entities.Budget) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.insert")
	defer span.End()

	query := `insert into
				budgets (
					id,
					user_id,
					date,
					amount_goal,
					amount_used,
					percentage_used,
					created_at,
					updated_at,
					deleted_at
					)
			  values
				($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.exec.ExecContext(
		ctx,
		query,
		budget.ID.Value,
		budget.UserID.Value,
		budget.Date,
		budget.AmountGoal.Money(),
		budget.AmountUsed.Money(),
		budget.PercentageUsed.Percentage(),
		budget.CreatedAt,
		budget.UpdatedAt.Time,
		budget.DeletedAt.Time,
	)
	if err != nil {
		span.AddEvent("error inserting budget", o11y.Attribute{Key: "budget_id", Value: budget.ID.Value}, o11y.Attribute{Key: "error", Value: err})
		r.o11y.Logger().Error(ctx, err, "error inserting budget", o11y.Field{Key: "budget_id", Value: budget.ID.Value})
		return err
	}
	return nil
}

func (r *budgetRepository) InsertItems(ctx context.Context, items []*entities.BudgetItem) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.insert_items")
	defer span.End()

	if len(items) == 0 {
		return nil
	}

	// Build batch insert query with multiple VALUES clauses
	const numColumns = 11
	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]any, 0, len(items)*numColumns)

	for i, item := range items {
		placeholderStart := i*numColumns + 1
		placeholders := make([]string, numColumns)
		for j := range numColumns {
			placeholders[j] = fmt.Sprintf("$%d", placeholderStart+j)
		}
		valueStrings = append(valueStrings, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))

		valueArgs = append(valueArgs,
			item.ID.Value,
			item.BudgetID.Value,
			item.CategoryID.Value,
			item.PercentageGoal.Percentage(),
			item.AmountGoal.Money(),
			item.AmountUsed.Money(),
			item.PercentageUsed.Percentage(),
			item.PercentageTotal.Percentage(),
			item.CreatedAt,
			item.UpdatedAt.Time,
			item.DeletedAt.Time,
		)
	}

	query := fmt.Sprintf(`insert into
				budget_items (
					id,
					budget_id,
					category_id,
					percentage_goal,
					amount_goal,
					amount_used,
					percentage_used,
					percentage_total,
					created_at,
					updated_at,
					deleted_at
				)
				values %s`, strings.Join(valueStrings, ", "))

	_, err := r.exec.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		span.AddEvent("error batch inserting budget items", o11y.Attribute{Key: "count", Value: len(items)}, o11y.Attribute{Key: "error", Value: err})
		r.o11y.Logger().Error(ctx, err, "error batch inserting budget items", o11y.Field{Key: "count", Value: len(items)})
		return err
	}

	return nil
}
