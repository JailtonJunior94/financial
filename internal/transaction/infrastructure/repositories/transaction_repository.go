package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type transactionRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

// NewTransactionRepository cria uma nova instância do repositório.
func NewTransactionRepository(
	db database.DBTX,
	o11y observability.Observability,
) interfaces.TransactionRepository {
	return &transactionRepository{
		db:   db,
		o11y: o11y,
	}
}

// FindOrCreateMonthly busca ou cria o aggregate do mês.
func (r *transactionRepository) FindOrCreateMonthly(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	referenceMonth transactionVos.ReferenceMonth,
) (*entities.MonthlyTransaction, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_or_create_monthly")
	defer span.End()

	// Tenta buscar existente
	monthly, err := r.findMonthlyByUserAndMonth(ctx, executor, userID, referenceMonth)
	if err != nil {
		return nil, err
	}

	// Se encontrou, retorna
	if monthly != nil {
		// Carrega os items
		items, err := r.findItemsByMonthlyID(ctx, executor, monthly.ID)
		if err != nil {
			return nil, err
		}
		monthly.LoadItems(items)
		return monthly, nil
	}

	// Se não encontrou, cria novo
	monthly, err = entities.NewMonthlyTransaction(userID, referenceMonth)
	if err != nil {
		return nil, err
	}

	// Gera ID
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	monthly.SetID(id)

	// Persiste
	if err := r.insertMonthly(ctx, executor, monthly); err != nil {
		return nil, err
	}

	return monthly, nil
}

// FindMonthlyByID busca o aggregate por ID com todos os items.
func (r *transactionRepository) FindMonthlyByID(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	monthlyID sharedVos.UUID,
) (*entities.MonthlyTransaction, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_monthly_by_id")
	defer span.End()

	query := `
		SELECT 
			id, user_id, reference_month, 
			total_income, total_expense, total_amount,
			created_at, updated_at
		FROM monthly_transactions
		WHERE id = $1 AND user_id = $2
	`

	var monthly entities.MonthlyTransaction
	var refMonthStr string
	var totalIncomeInt, totalExpenseInt, totalAmountInt int64
	var createdAt time.Time
	var updatedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, monthlyID.String(), userID.String()).Scan(
		&monthly.ID,
		&monthly.UserID,
		&refMonthStr,
		&totalIncomeInt,
		&totalExpenseInt,
		&totalAmountInt,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to find monthly transaction", observability.Error(err))
		return nil, err
	}

	// Parse reference month
	monthly.ReferenceMonth, err = transactionVos.NewReferenceMonthFromString(refMonthStr)
	if err != nil {
		return nil, err
	}

	// Parse money values
	monthly.TotalIncome, err = sharedVos.NewMoney(totalIncomeInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	monthly.TotalExpense, err = sharedVos.NewMoney(totalExpenseInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	monthly.TotalAmount, err = sharedVos.NewMoney(totalAmountInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	// Parse timestamps
	monthly.CreatedAt = sharedVos.NewNullableTime(createdAt)
	if updatedAt.Valid {
		monthly.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
	}

	// Load items
	items, err := r.findItemsByMonthlyID(ctx, executor, monthly.ID)
	if err != nil {
		return nil, err
	}
	monthly.LoadItems(items)

	return &monthly, nil
}

// UpdateMonthly atualiza o aggregate (totais).
func (r *transactionRepository) UpdateMonthly(
	ctx context.Context,
	executor database.DBTX,
	monthly *entities.MonthlyTransaction,
) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.update_monthly")
	defer span.End()

	query := `
		UPDATE monthly_transactions
		SET 
			total_income = $1,
			total_expense = $2,
			total_amount = $3,
			updated_at = $4
		WHERE id = $5
	`

	_, err := executor.ExecContext(ctx, query,
		monthly.TotalIncome.Cents(),
		monthly.TotalExpense.Cents(),
		monthly.TotalAmount.Cents(),
		time.Now().UTC(),
		monthly.ID.String(),
	)

	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to update monthly transaction", observability.Error(err))
		return err
	}

	return nil
}

// InsertItem insere um novo transaction item.
func (r *transactionRepository) InsertItem(
	ctx context.Context,
	executor database.DBTX,
	item *entities.TransactionItem,
) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.insert_item")
	defer span.End()

	query := `
		INSERT INTO transaction_items (
			id, monthly_id, category_id, title, description,
			amount, direction, type, is_paid, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := executor.ExecContext(ctx, query,
		item.ID.String(),
		item.MonthlyID.String(),
		item.CategoryID.String(),
		item.Title,
		item.Description,
		item.Amount.Cents(),
		item.Direction.String(),
		item.Type.String(),
		item.IsPaid,
		item.CreatedAt.ValueOr(time.Now().UTC()),
	)

	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to insert transaction item", observability.Error(err))
		return err
	}

	return nil
}

// UpdateItem atualiza um transaction item existente.
func (r *transactionRepository) UpdateItem(
	ctx context.Context,
	executor database.DBTX,
	item *entities.TransactionItem,
) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.update_item")
	defer span.End()

	query := `
		UPDATE transaction_items
		SET 
			title = $1,
			description = $2,
			amount = $3,
			direction = $4,
			type = $5,
			is_paid = $6,
			updated_at = $7,
			deleted_at = $8
		WHERE id = $9
	`

	var deletedAt any
	if item.IsDeleted() {
		deletedAt = item.DeletedAt.ValueOr(time.Now().UTC())
	}

	_, err := executor.ExecContext(ctx, query,
		item.Title,
		item.Description,
		item.Amount.Cents(),
		item.Direction.String(),
		item.Type.String(),
		item.IsPaid,
		time.Now().UTC(),
		deletedAt,
		item.ID.String(),
	)

	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to update transaction item", observability.Error(err))
		return err
	}

	return nil
}

// FindItemByID busca um item por ID.
func (r *transactionRepository) FindItemByID(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	itemID sharedVos.UUID,
) (*entities.TransactionItem, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_item_by_id")
	defer span.End()

	query := `
		SELECT 
			ti.id, ti.monthly_id, ti.category_id, ti.title, ti.description,
			ti.amount, ti.direction, ti.type, ti.is_paid,
			ti.created_at, ti.updated_at, ti.deleted_at
		FROM transaction_items ti
		INNER JOIN monthly_transactions mt ON ti.monthly_id = mt.id
		WHERE ti.id = $1 AND mt.user_id = $2
	`

	var item entities.TransactionItem
	var amountInt int64
	var direction, itemType string
	var createdAt time.Time
	var updatedAt, deletedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, itemID.String(), userID.String()).Scan(
		&item.ID,
		&item.MonthlyID,
		&item.CategoryID,
		&item.Title,
		&item.Description,
		&amountInt,
		&direction,
		&itemType,
		&item.IsPaid,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to find transaction item", observability.Error(err))
		return nil, err
	}

	// Parse values
	item.Amount, err = sharedVos.NewMoney(amountInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	item.Direction, err = transactionVos.NewTransactionDirection(direction)
	if err != nil {
		return nil, err
	}

	item.Type, err = transactionVos.NewTransactionType(itemType)
	if err != nil {
		return nil, err
	}

	item.CreatedAt = sharedVos.NewNullableTime(createdAt)
	if updatedAt.Valid {
		item.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
	}
	if deletedAt.Valid {
		item.DeletedAt = sharedVos.NewNullableTime(deletedAt.Time)
	}

	return &item, nil
}

// --- Private methods ---

// findMonthlyByUserAndMonth busca aggregate por user e mês.
func (r *transactionRepository) findMonthlyByUserAndMonth(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	referenceMonth transactionVos.ReferenceMonth,
) (*entities.MonthlyTransaction, error) {
	query := `
		SELECT 
			id, user_id, reference_month,
			total_income, total_expense, total_amount,
			created_at, updated_at
		FROM monthly_transactions
		WHERE user_id = $1 AND reference_month = $2
	`

	var monthly entities.MonthlyTransaction
	var refMonthStr string
	var totalIncomeInt, totalExpenseInt, totalAmountInt int64
	var createdAt time.Time
	var updatedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, userID.String(), referenceMonth.String()).Scan(
		&monthly.ID,
		&monthly.UserID,
		&refMonthStr,
		&totalIncomeInt,
		&totalExpenseInt,
		&totalAmountInt,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse values
	monthly.ReferenceMonth, err = transactionVos.NewReferenceMonthFromString(refMonthStr)
	if err != nil {
		return nil, err
	}

	monthly.TotalIncome, err = sharedVos.NewMoney(totalIncomeInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	monthly.TotalExpense, err = sharedVos.NewMoney(totalExpenseInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	monthly.TotalAmount, err = sharedVos.NewMoney(totalAmountInt, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, err
	}

	monthly.CreatedAt = sharedVos.NewNullableTime(createdAt)
	if updatedAt.Valid {
		monthly.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
	}

	return &monthly, nil
}

// insertMonthly insere um novo aggregate.
func (r *transactionRepository) insertMonthly(
	ctx context.Context,
	executor database.DBTX,
	monthly *entities.MonthlyTransaction,
) error {
	query := `
		INSERT INTO monthly_transactions (
			id, user_id, reference_month,
			total_income, total_expense, total_amount,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := executor.ExecContext(ctx, query,
		monthly.ID.String(),
		monthly.UserID.String(),
		monthly.ReferenceMonth.String(),
		monthly.TotalIncome.Cents(),
		monthly.TotalExpense.Cents(),
		monthly.TotalAmount.Cents(),
		monthly.CreatedAt.ValueOr(time.Now().UTC()),
	)

	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to insert monthly transaction", observability.Error(err))
		return err
	}

	return nil
}

// findItemsByMonthlyID busca todos os items de um aggregate (exceto deletados).
func (r *transactionRepository) findItemsByMonthlyID(
	ctx context.Context,
	executor database.DBTX,
	monthlyID sharedVos.UUID,
) ([]*entities.TransactionItem, error) {
	query := `
		SELECT 
			id, monthly_id, category_id, title, description,
			amount, direction, type, is_paid,
			created_at, updated_at, deleted_at
		FROM transaction_items
		WHERE monthly_id = $1
		ORDER BY created_at ASC
	`

	rows, err := executor.QueryContext(ctx, query, monthlyID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.o11y.Logger().Error(ctx, "failed to close rows in findItemsByMonthlyID",
				observability.Error(closeErr),
			)
		}
	}()

	items := make([]*entities.TransactionItem, 0)

	for rows.Next() {
		var item entities.TransactionItem
		var amountInt int64
		var direction, itemType string
		var createdAt time.Time
		var updatedAt, deletedAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.MonthlyID,
			&item.CategoryID,
			&item.Title,
			&item.Description,
			&amountInt,
			&direction,
			&itemType,
			&item.IsPaid,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse values
		item.Amount, err = sharedVos.NewMoney(amountInt, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, err
		}

		item.Direction, err = transactionVos.NewTransactionDirection(direction)
		if err != nil {
			return nil, err
		}

		item.Type, err = transactionVos.NewTransactionType(itemType)
		if err != nil {
			return nil, err
		}

		item.CreatedAt = sharedVos.NewNullableTime(createdAt)
		if updatedAt.Valid {
			item.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
		}
		if deletedAt.Valid {
			item.DeletedAt = sharedVos.NewNullableTime(deletedAt.Time)
		}

		items = append(items, &item)
	}

	return items, rows.Err()
}

// ListMonthlyPaginated lista monthly transactions com paginação cursor-based.
func (r *transactionRepository) ListMonthlyPaginated(
	ctx context.Context,
	params interfaces.ListMonthlyParams,
) ([]*entities.MonthlyTransaction, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.list_monthly_paginated")
	defer span.End()

	// Build WHERE clause with cursor for keyset pagination
	whereClause := "user_id = $1"
	args := []interface{}{params.UserID.String()}

	cursorDate, hasDate := params.Cursor.GetString("date")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasDate && hasID && cursorDate != "" && cursorID != "" {
		whereClause += ` AND (
			reference_month < $2
			OR (reference_month = $2 AND id < $3)
		)`
		args = append(args, cursorDate, cursorID)
	}

	query := fmt.Sprintf(`
		SELECT
			id, user_id, reference_month,
			total_income, total_expense, total_amount,
			created_at, updated_at
		FROM monthly_transactions
		WHERE %s
		ORDER BY reference_month DESC, id DESC
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to list monthly transactions", observability.Error(err))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var monthlyList []*entities.MonthlyTransaction

	for rows.Next() {
		var monthly entities.MonthlyTransaction
		var refMonthStr string
		var totalIncomeInt, totalExpenseInt, totalAmountInt int64
		var createdAt time.Time
		var updatedAt sql.NullTime

		err := rows.Scan(
			&monthly.ID,
			&monthly.UserID,
			&refMonthStr,
			&totalIncomeInt,
			&totalExpenseInt,
			&totalAmountInt,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			r.o11y.Logger().Error(ctx, "failed to scan monthly transaction", observability.Error(err))
			return nil, err
		}

		// Parse reference month
		monthly.ReferenceMonth, err = transactionVos.NewReferenceMonthFromString(refMonthStr)
		if err != nil {
			return nil, err
		}

		// Parse money values
		monthly.TotalIncome, err = sharedVos.NewMoney(totalIncomeInt, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, err
		}

		monthly.TotalExpense, err = sharedVos.NewMoney(totalExpenseInt, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, err
		}

		monthly.TotalAmount, err = sharedVos.NewMoney(totalAmountInt, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, err
		}

		// Parse timestamps
		monthly.CreatedAt = sharedVos.NewNullableTime(createdAt)
		if updatedAt.Valid {
			monthly.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
		}

		// Load items for this monthly transaction
		items, err := r.findItemsByMonthlyID(ctx, r.db, monthly.ID)
		if err != nil {
			return nil, err
		}
		monthly.LoadItems(items)

		monthlyList = append(monthlyList, &monthly)
	}

	return monthlyList, rows.Err()
}

// GetMonthlyByID busca um monthly transaction por ID (sem executor UoW).
func (r *transactionRepository) GetMonthlyByID(
	ctx context.Context,
	userID sharedVos.UUID,
	monthlyID sharedVos.UUID,
) (*entities.MonthlyTransaction, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.get_monthly_by_id")
	defer span.End()

	return r.FindMonthlyByID(ctx, r.db, userID, monthlyID)
}
