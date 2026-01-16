package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type budgetRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewBudgetRepository(db database.DBTX, o11y observability.Observability) interfaces.BudgetRepository {
	return &budgetRepository{
		db:   db,
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

	_, err := r.db.ExecContext(
		ctx,
		query,
		budget.ID.Value,
		budget.UserID.Value,
		budget.ReferenceMonth.ToTime(),
		budget.TotalAmount,
		budget.SpentAmount,
		budget.PercentageUsed,
		budget.CreatedAt,
		budget.UpdatedAt.Ptr(),
		budget.DeletedAt.Ptr(),
	)
	if err != nil {
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
	const numColumns = 9
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
			item.PercentageGoal,
			item.PlannedAmount,
			item.SpentAmount,
			item.CreatedAt,
			item.UpdatedAt.Ptr(),
			item.DeletedAt.Ptr(),
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
					created_at,
					updated_at,
					deleted_at
				)
				values %s`, strings.Join(valueStrings, ", "))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *budgetRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Budget, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.find_by_id")
	defer span.End()

	query := `select
				b.id,
				b.user_id,
				b.date,
				b.amount_goal,
				b.amount_used,
				b.percentage_used,
				b.created_at,
				b.updated_at,
				b.deleted_at
			from budgets b
			where b.id = $1 and b.deleted_at is null`

	row := r.db.QueryRowContext(ctx, query, id.Value)

	var budget entities.Budget
	var updatedAt, deletedAt *time.Time
	var amountGoal, amountUsed, percentageUsed string
	var referenceDate time.Time

	err := row.Scan(
		&budget.ID.Value,
		&budget.UserID.Value,
		&referenceDate,
		&amountGoal,
		&amountUsed,
		&percentageUsed,
		&budget.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse NUMERIC values from strings
	amountGoalFloat, err := strconv.ParseFloat(amountGoal, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_goal: %w", err)
	}

	amountUsedFloat, err := strconv.ParseFloat(amountUsed, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_used: %w", err)
	}

	percentageUsedFloat, err := strconv.ParseFloat(percentageUsed, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse percentage_used: %w", err)
	}

	budget.TotalAmount, _ = vos.NewMoneyFromFloat(amountGoalFloat, "BRL")
	budget.SpentAmount, _ = vos.NewMoneyFromFloat(amountUsedFloat, "BRL")
	budget.PercentageUsed, _ = vos.NewPercentageFromFloat(percentageUsedFloat)
	budget.ReferenceMonth = budgetVos.NewReferenceMonthFromDate(referenceDate)

	if updatedAt != nil {
		budget.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		budget.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	// Load budget items
	items, err := r.findItemsByBudgetID(ctx, id)
	if err != nil {
		return nil, err
	}
	budget.Items = items

	return &budget, nil
}

func (r *budgetRepository) FindByUserIDAndReferenceMonth(ctx context.Context, userID vos.UUID, referenceMonth budgetVos.ReferenceMonth) (*entities.Budget, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.find_by_user_and_month")
	defer span.End()

	query := `select
				b.id,
				b.user_id,
				b.date,
				b.amount_goal,
				b.amount_used,
				b.percentage_used,
				b.created_at,
				b.updated_at,
				b.deleted_at
			from budgets b
			where b.user_id = $1
			  and to_char(b.date, 'YYYY-MM') = $2
			  and b.deleted_at is null`

	row := r.db.QueryRowContext(ctx, query, userID.Value, referenceMonth.String())

	var budget entities.Budget
	var updatedAt, deletedAt *time.Time
	var amountGoal, amountUsed, percentageUsed string
	var referenceDate time.Time

	err := row.Scan(
		&budget.ID.Value,
		&budget.UserID.Value,
		&referenceDate,
		&amountGoal,
		&amountUsed,
		&percentageUsed,
		&budget.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse NUMERIC values from strings
	amountGoalFloat, err := strconv.ParseFloat(amountGoal, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_goal: %w", err)
	}

	amountUsedFloat, err := strconv.ParseFloat(amountUsed, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_used: %w", err)
	}

	percentageUsedFloat, err := strconv.ParseFloat(percentageUsed, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse percentage_used: %w", err)
	}

	budget.TotalAmount, _ = vos.NewMoneyFromFloat(amountGoalFloat, "BRL")
	budget.SpentAmount, _ = vos.NewMoneyFromFloat(amountUsedFloat, "BRL")
	budget.PercentageUsed, _ = vos.NewPercentageFromFloat(percentageUsedFloat)
	budget.ReferenceMonth = budgetVos.NewReferenceMonthFromDate(referenceDate)

	if updatedAt != nil {
		budget.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		budget.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	// Load budget items
	items, err := r.findItemsByBudgetID(ctx, budget.ID)
	if err != nil {
		return nil, err
	}
	budget.Items = items

	return &budget, nil
}

func (r *budgetRepository) Update(ctx context.Context, budget *entities.Budget) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.update")
	defer span.End()

	query := `update budgets set
				amount_goal = $2,
				amount_used = $3,
				percentage_used = $4,
				updated_at = $5
			where id = $1`

	_, err := r.db.ExecContext(
		ctx,
		query,
		budget.ID.Value,
		budget.TotalAmount,
		budget.SpentAmount,
		budget.PercentageUsed,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *budgetRepository) UpdateItem(ctx context.Context, item *entities.BudgetItem) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "budget_repository.update_item")
	defer span.End()

	query := `update budget_items set
				amount_used = $2,
				updated_at = $3
			where id = $1`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID.Value,
		float64(item.SpentAmount.Cents())/100.0,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *budgetRepository) findItemsByBudgetID(ctx context.Context, budgetID vos.UUID) ([]*entities.BudgetItem, error) {
	query := `select
				id,
				budget_id,
				category_id,
				percentage_goal,
				amount_goal,
				amount_used,
				created_at,
				updated_at,
				deleted_at
			from budget_items
			where budget_id = $1 and deleted_at is null
			order by created_at`

	rows, err := r.db.QueryContext(ctx, query, budgetID.Value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entities.BudgetItem
	for rows.Next() {
		var item entities.BudgetItem
		var updatedAt, deletedAt *time.Time
		var amountGoal, amountUsed, percentageGoal string

		err := rows.Scan(
			&item.ID.Value,
			&item.BudgetID.Value,
			&item.CategoryID.Value,
			&percentageGoal,
			&amountGoal,
			&amountUsed,
			&item.CreatedAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse NUMERIC values from strings
		percentageGoalFloat, err := strconv.ParseFloat(percentageGoal, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse percentage_goal: %w", err)
		}

		amountGoalFloat, err := strconv.ParseFloat(amountGoal, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount_goal: %w", err)
		}

		amountUsedFloat, err := strconv.ParseFloat(amountUsed, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount_used: %w", err)
		}

		item.PlannedAmount, _ = vos.NewMoneyFromFloat(amountGoalFloat, "BRL")
		item.SpentAmount, _ = vos.NewMoneyFromFloat(amountUsedFloat, "BRL")
		item.PercentageGoal, _ = vos.NewPercentageFromFloat(percentageGoalFloat)

		if updatedAt != nil {
			item.UpdatedAt = vos.NewNullableTime(*updatedAt)
		}
		if deletedAt != nil {
			item.DeletedAt = vos.NewNullableTime(*deletedAt)
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}
