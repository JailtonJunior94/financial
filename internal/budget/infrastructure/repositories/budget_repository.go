package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type budgetRepository struct {
	db   *sql.DB
	tx   *sql.Tx
	o11y o11y.Observability
}

func NewBudgetRepository(db *sql.DB, tx *sql.Tx, o11y o11y.Observability) interfaces.BudgetRepository {
	return &budgetRepository{
		db:   db,
		tx:   tx,
		o11y: o11y,
	}
}

func (r *budgetRepository) Insert(ctx context.Context, budget *entities.Budget) error {
	ctx, span := r.o11y.Start(ctx, "budget_repository.insert")
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
				(?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.tx.ExecContext(
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
		span.AddAttributes(ctx, o11y.Error, "error insert budget", o11y.Attributes{Key: "error", Value: err})
		return err
	}
	return nil
}

func (r *budgetRepository) InsertItems(ctx context.Context, items []*entities.BudgetItem) error {
	ctx, span := r.o11y.Start(ctx, "budget_repository.insert_items")
	defer span.End()

	query := `insert into
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
				values
					(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, item := range items {
		_, err := r.tx.ExecContext(
			ctx,
			query,
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
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert budget item", o11y.Attributes{Key: "error", Value: err})
			return err
		}
	}

	return nil
}
