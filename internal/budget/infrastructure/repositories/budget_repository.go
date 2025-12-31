package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/database"
)

type budgetRepository struct {
	exec database.DBExecutor
	o11y observability.Observability
}

func NewBudgetRepository(exec database.DBExecutor, o11y observability.Observability) interfaces.BudgetRepository {
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
		budget.ReferenceMonth.ToTime(),
		budget.TotalAmount.Cents(),
		budget.SpentAmount.Cents(),
		budget.PercentageUsed.Value(),
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

		percentageSpent := item.PercentageSpent()
		valueArgs = append(valueArgs,
			item.ID.Value,
			item.BudgetID.Value,
			item.CategoryID.Value,
			item.PercentageGoal.Value(),
			item.PlannedAmount.Cents(),
			item.SpentAmount.Cents(),
			percentageSpent.Value(),
			percentageSpent.Value(),
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
					percentage_used,
					percentage_total,
					created_at,
					updated_at,
					deleted_at
				)
				values %s`, strings.Join(valueStrings, ", "))

	_, err := r.exec.ExecContext(ctx, query, valueArgs...)
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

	row := r.exec.QueryRowContext(ctx, query, id.Value)

	var budget entities.Budget
	var updatedAt, deletedAt *time.Time
	var amountGoalCents, amountUsedCents int64
	var percentageUsedValue int64
	var referenceDate time.Time

	err := row.Scan(
		&budget.ID.Value,
		&budget.UserID.Value,
		&referenceDate,
		&amountGoalCents,
		&amountUsedCents,
		&percentageUsedValue,
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

	// Reconstruct Money and Percentage types
	totalAmountFloat := float64(amountGoalCents) / 100.0
	spentAmountFloat := float64(amountUsedCents) / 100.0

	budget.TotalAmount, _ = vos.NewMoneyFromFloat(totalAmountFloat, "BRL")
	budget.SpentAmount, _ = vos.NewMoneyFromFloat(spentAmountFloat, "BRL")
	budget.PercentageUsed, _ = vos.NewPercentage(percentageUsedValue)
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

	row := r.exec.QueryRowContext(ctx, query, userID.Value, referenceMonth.String())

	var budget entities.Budget
	var updatedAt, deletedAt *time.Time
	var amountGoalCents, amountUsedCents int64
	var percentageUsedValue int64
	var referenceDate time.Time

	err := row.Scan(
		&budget.ID.Value,
		&budget.UserID.Value,
		&referenceDate,
		&amountGoalCents,
		&amountUsedCents,
		&percentageUsedValue,
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

	// Reconstruct Money and Percentage types
	totalAmountFloat := float64(amountGoalCents) / 100.0
	spentAmountFloat := float64(amountUsedCents) / 100.0

	budget.TotalAmount, _ = vos.NewMoneyFromFloat(totalAmountFloat, "BRL")
	budget.SpentAmount, _ = vos.NewMoneyFromFloat(spentAmountFloat, "BRL")
	budget.PercentageUsed, _ = vos.NewPercentage(percentageUsedValue)
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

	_, err := r.exec.ExecContext(
		ctx,
		query,
		budget.ID.Value,
		budget.TotalAmount.Cents(),
		budget.SpentAmount.Cents(),
		budget.PercentageUsed.Value(),
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
				percentage_used = $3,
				percentage_total = $4,
				updated_at = $5
			where id = $1`

	percentageUsed := item.PercentageSpent()

	_, err := r.exec.ExecContext(
		ctx,
		query,
		item.ID.Value,
		item.SpentAmount.Cents(),
		percentageUsed.Value(),
		percentageUsed.Value(), // percentage_total is same as percentage_used for item
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
				percentage_used,
				percentage_total,
				created_at,
				updated_at,
				deleted_at
			from budget_items
			where budget_id = $1 and deleted_at is null
			order by created_at`

	rows, err := r.exec.QueryContext(ctx, query, budgetID.Value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entities.BudgetItem
	for rows.Next() {
		var item entities.BudgetItem
		var updatedAt, deletedAt *time.Time
		var amountGoalCents, amountUsedCents int64
		var percentageGoalValue, percentageUsedValue, percentageTotalValue int64

		err := rows.Scan(
			&item.ID.Value,
			&item.BudgetID.Value,
			&item.CategoryID.Value,
			&percentageGoalValue,
			&amountGoalCents,
			&amountUsedCents,
			&percentageUsedValue,
			&percentageTotalValue,
			&item.CreatedAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Reconstruct Money and Percentage types
		plannedAmountFloat := float64(amountGoalCents) / 100.0
		spentAmountFloat := float64(amountUsedCents) / 100.0

		item.PlannedAmount, _ = vos.NewMoneyFromFloat(plannedAmountFloat, "BRL")
		item.SpentAmount, _ = vos.NewMoneyFromFloat(spentAmountFloat, "BRL")
		item.PercentageGoal, _ = vos.NewPercentage(percentageGoalValue)

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
