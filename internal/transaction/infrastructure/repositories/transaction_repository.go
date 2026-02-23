package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"

	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type transactionRepository struct {
	db   database.DBTX
	o11y observability.Observability
	fm   *metrics.FinancialMetrics
}

// NewTransactionRepository cria uma nova instância do repositório.
func NewTransactionRepository(
	db database.DBTX,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) interfaces.TransactionRepository {
	return &transactionRepository{
		db:   db,
		o11y: o11y,
		fm:   fm,
	}
}

// FindOrCreateMonthly busca ou cria o aggregate do mês.
func (r *transactionRepository) FindOrCreateMonthly(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
) (*entities.MonthlyTransaction, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_or_create_monthly")
	defer span.End()

	// Tenta buscar existente
	monthly, err := r.findMonthlyByUserAndMonth(ctx, executor, userID, referenceMonth)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_or_create_monthly", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Se encontrou, retorna
	if monthly != nil {
		// Carrega os items
		items, err := r.findItemsByMonthlyID(ctx, executor, monthly.ID)
		if err != nil {
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "find_or_create_monthly", "transaction", "infra", time.Since(start))
			return nil, err
		}
		monthly.LoadItems(items)
		r.fm.RecordRepositoryQuery(ctx, "find_or_create_monthly", "transaction", time.Since(start))
		return monthly, nil
	}

	// Se não encontrou, cria novo
	monthly, err = entities.NewMonthlyTransaction(userID, referenceMonth)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_or_create_monthly", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Gera ID
	id, err := sharedVos.NewUUID()
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_or_create_monthly", "transaction", "infra", time.Since(start))
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	monthly.SetID(id)

	// Persiste
	if err := r.insertMonthly(ctx, executor, monthly); err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_or_create_monthly", "transaction", "infra", time.Since(start))
		return nil, err
	}

	r.fm.RecordRepositoryQuery(ctx, "find_or_create_monthly", "transaction", time.Since(start))
	return monthly, nil
}

// FindMonthlyByID busca o aggregate por ID com todos os items.
func (r *transactionRepository) FindMonthlyByID(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	monthlyID sharedVos.UUID,
) (*entities.MonthlyTransaction, error) {
	start := time.Now()
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
	var totalIncomeStr, totalExpenseStr, totalAmountStr string
	var createdAt time.Time
	var updatedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, monthlyID.String(), userID.String()).Scan(
		&monthly.ID.Value,
		&monthly.UserID.Value,
		&refMonthStr,
		&totalIncomeStr,
		&totalExpenseStr,
		&totalAmountStr,
		&createdAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		r.fm.RecordRepositoryQuery(ctx, "find_monthly_by_id", "transaction", time.Since(start))
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to find monthly transaction", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Parse reference month
	monthly.ReferenceMonth, err = pkgVos.NewReferenceMonth(refMonthStr)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Parse money values from NUMERIC strings
	monthly.TotalIncome, err = sharedVos.NewMoneyFromString(totalIncomeStr, sharedVos.CurrencyBRL)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, fmt.Errorf("failed to parse total_income: %w", err)
	}

	monthly.TotalExpense, err = sharedVos.NewMoneyFromString(totalExpenseStr, sharedVos.CurrencyBRL)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, fmt.Errorf("failed to parse total_expense: %w", err)
	}

	monthly.TotalAmount, err = sharedVos.NewMoneyFromString(totalAmountStr, sharedVos.CurrencyBRL)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, fmt.Errorf("failed to parse total_amount: %w", err)
	}

	// Parse timestamps
	monthly.CreatedAt = sharedVos.NewNullableTime(createdAt)
	if updatedAt.Valid {
		monthly.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
	}

	// Load items
	items, err := r.findItemsByMonthlyID(ctx, executor, monthly.ID)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_monthly_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}
	monthly.LoadItems(items)

	r.fm.RecordRepositoryQuery(ctx, "find_monthly_by_id", "transaction", time.Since(start))
	return &monthly, nil
}

// UpdateMonthly atualiza o aggregate (totais).
func (r *transactionRepository) UpdateMonthly(
	ctx context.Context,
	executor database.DBTX,
	monthly *entities.MonthlyTransaction,
) error {
	start := time.Now()
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

	now := sharedVos.NewNullableTime(time.Now().UTC())

	_, err := executor.ExecContext(ctx, query,
		monthly.TotalIncome.Float(),
		monthly.TotalExpense.Float(),
		monthly.TotalAmount.Float(),
		now.ValueOr(time.Now().UTC()),
		monthly.ID.String(),
	)

	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to update monthly transaction", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "update_monthly", "transaction", "infra", time.Since(start))
		return err
	}

	r.fm.RecordRepositoryQuery(ctx, "update_monthly", "transaction", time.Since(start))
	return nil
}

// InsertItem insere um novo transaction item.
func (r *transactionRepository) InsertItem(
	ctx context.Context,
	executor database.DBTX,
	item *entities.TransactionItem,
) error {
	start := time.Now()
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
		item.Amount.Float(),
		item.Direction.String(),
		item.Type.String(),
		item.IsPaid,
		item.CreatedAt.ValueOr(time.Now().UTC()),
	)

	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to insert transaction item", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "insert_item", "transaction", "infra", time.Since(start))
		return err
	}

	r.fm.RecordRepositoryQuery(ctx, "insert_item", "transaction", time.Since(start))
	return nil
}

// UpdateItem atualiza um transaction item existente.
func (r *transactionRepository) UpdateItem(
	ctx context.Context,
	executor database.DBTX,
	item *entities.TransactionItem,
) error {
	start := time.Now()
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

	now := sharedVos.NewNullableTime(time.Now().UTC())

	var deletedAt any
	if item.IsDeleted() {
		deletedAt = item.DeletedAt.ValueOr(now.ValueOr(time.Now().UTC()))
	}

	_, err := executor.ExecContext(ctx, query,
		item.Title,
		item.Description,
		item.Amount.Float(),
		item.Direction.String(),
		item.Type.String(),
		item.IsPaid,
		now.ValueOr(time.Now().UTC()),
		deletedAt,
		item.ID.String(),
	)

	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to update transaction item", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "update_item", "transaction", "infra", time.Since(start))
		return err
	}

	r.fm.RecordRepositoryQuery(ctx, "update_item", "transaction", time.Since(start))
	return nil
}

// FindItemByID busca um item por ID.
func (r *transactionRepository) FindItemByID(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	itemID sharedVos.UUID,
) (*entities.TransactionItem, error) {
	start := time.Now()
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
	var amountStr string
	var direction, itemType string
	var createdAt time.Time
	var updatedAt, deletedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, itemID.String(), userID.String()).Scan(
		&item.ID.Value,
		&item.MonthlyID.Value,
		&item.CategoryID.Value,
		&item.Title,
		&item.Description,
		&amountStr,
		&direction,
		&itemType,
		&item.IsPaid,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		r.fm.RecordRepositoryQuery(ctx, "find_item_by_id", "transaction", time.Since(start))
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to find transaction item", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "find_item_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Parse values
	item.Amount, err = sharedVos.NewMoneyFromString(amountStr, sharedVos.CurrencyBRL)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_item_by_id", "transaction", "infra", time.Since(start))
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	item.Direction, err = transactionVos.NewTransactionDirection(direction)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_item_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	item.Type, err = transactionVos.NewTransactionType(itemType)
	if err != nil {
		r.fm.RecordRepositoryFailure(ctx, "find_item_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	item.CreatedAt = sharedVos.NewNullableTime(createdAt)
	if updatedAt.Valid {
		item.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
	}
	if deletedAt.Valid {
		item.DeletedAt = sharedVos.NewNullableTime(deletedAt.Time)
	}

	r.fm.RecordRepositoryQuery(ctx, "find_item_by_id", "transaction", time.Since(start))
	return &item, nil
}

// --- Private methods ---

// findMonthlyByUserAndMonth busca aggregate por user e mês.
func (r *transactionRepository) findMonthlyByUserAndMonth(
	ctx context.Context,
	executor database.DBTX,
	userID sharedVos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
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
	var totalIncomeStr, totalExpenseStr, totalAmountStr string
	var createdAt time.Time
	var updatedAt sql.NullTime

	err := executor.QueryRowContext(ctx, query, userID.String(), referenceMonth.String()).Scan(
		&monthly.ID.Value,
		&monthly.UserID.Value,
		&refMonthStr,
		&totalIncomeStr,
		&totalExpenseStr,
		&totalAmountStr,
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
	monthly.ReferenceMonth, err = pkgVos.NewReferenceMonth(refMonthStr)
	if err != nil {
		return nil, err
	}

	// Parse money values from NUMERIC strings
	monthly.TotalIncome, err = sharedVos.NewMoneyFromString(totalIncomeStr, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total_income: %w", err)
	}

	monthly.TotalExpense, err = sharedVos.NewMoneyFromString(totalExpenseStr, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total_expense: %w", err)
	}

	monthly.TotalAmount, err = sharedVos.NewMoneyFromString(totalAmountStr, sharedVos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total_amount: %w", err)
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
		monthly.TotalIncome.Float(),
		monthly.TotalExpense.Float(),
		monthly.TotalAmount.Float(),
		monthly.CreatedAt.ValueOr(time.Now().UTC()),
	)

	if err != nil {
		r.o11y.Logger().Error(ctx, "failed to insert monthly transaction", observability.Error(err))
		return err
	}

	return nil
}

// findItemsByMonthlyIDs busca items de múltiplos aggregates em uma única query (evita N+1).
func (r *transactionRepository) findItemsByMonthlyIDs(
	ctx context.Context,
	ids []sharedVos.UUID,
) (map[string][]*entities.TransactionItem, error) {
	if len(ids) == 0 {
		return map[string][]*entities.TransactionItem{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id.String()
	}

	query := fmt.Sprintf(`
		SELECT
			id, monthly_id, category_id, title, description,
			amount, direction, type, is_paid,
			created_at, updated_at, deleted_at
		FROM transaction_items
		WHERE monthly_id IN (%s)
		ORDER BY monthly_id, created_at ASC
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string][]*entities.TransactionItem)

	for rows.Next() {
		var item entities.TransactionItem
		var amountStr string
		var direction, itemType string
		var createdAt time.Time
		var updatedAt, deletedAt sql.NullTime

		err := rows.Scan(
			&item.ID.Value,
			&item.MonthlyID.Value,
			&item.CategoryID.Value,
			&item.Title,
			&item.Description,
			&amountStr,
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

		item.Amount, err = sharedVos.NewMoneyFromString(amountStr, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
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

		key := item.MonthlyID.String()
		result[key] = append(result[key], &item)
	}

	return result, rows.Err()
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
		var amountStr string
		var direction, itemType string
		var createdAt time.Time
		var updatedAt, deletedAt sql.NullTime

		err := rows.Scan(
			&item.ID.Value,
			&item.MonthlyID.Value,
			&item.CategoryID.Value,
			&item.Title,
			&item.Description,
			&amountStr,
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

		// Parse values using Money VO
		item.Amount, err = sharedVos.NewMoneyFromString(amountStr, sharedVos.CurrencyBRL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
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
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.list_monthly_paginated")
	defer span.End()

	// Build WHERE clause with cursor for keyset pagination
	whereClause := "user_id = $1"
	args := []any{params.UserID.String()}

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
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "failed to list monthly transactions", observability.Error(err))
		r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var monthlyList []*entities.MonthlyTransaction

	for rows.Next() {
		var monthly entities.MonthlyTransaction
		var refMonthStr string
		var totalIncomeStr, totalExpenseStr, totalAmountStr string
		var createdAt time.Time
		var updatedAt sql.NullTime

		err := rows.Scan(
			&monthly.ID.Value,
			&monthly.UserID.Value,
			&refMonthStr,
			&totalIncomeStr,
			&totalExpenseStr,
			&totalAmountStr,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "failed to scan monthly transaction", observability.Error(err))
			r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
			return nil, err
		}

		// Parse reference month
		monthly.ReferenceMonth, err = pkgVos.NewReferenceMonth(refMonthStr)
		if err != nil {
			r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
			return nil, err
		}

		// Parse money values from NUMERIC strings
		monthly.TotalIncome, err = sharedVos.NewMoneyFromString(totalIncomeStr, sharedVos.CurrencyBRL)
		if err != nil {
			r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
			return nil, fmt.Errorf("failed to parse total_income: %w", err)
		}

		monthly.TotalExpense, err = sharedVos.NewMoneyFromString(totalExpenseStr, sharedVos.CurrencyBRL)
		if err != nil {
			r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
			return nil, fmt.Errorf("failed to parse total_expense: %w", err)
		}

		monthly.TotalAmount, err = sharedVos.NewMoneyFromString(totalAmountStr, sharedVos.CurrencyBRL)
		if err != nil {
			r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
			return nil, fmt.Errorf("failed to parse total_amount: %w", err)
		}

		// Parse timestamps
		monthly.CreatedAt = sharedVos.NewNullableTime(createdAt)
		if updatedAt.Valid {
			monthly.UpdatedAt = sharedVos.NewNullableTime(updatedAt.Time)
		}

		monthlyList = append(monthlyList, &monthly)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
		return nil, err
	}

	// Bulk-fetch items for all monthly transactions in a single query (avoids N+1)
	ids := make([]sharedVos.UUID, len(monthlyList))
	for i, m := range monthlyList {
		ids[i] = m.ID
	}

	itemsByMonthly, err := r.findItemsByMonthlyIDs(ctx, ids)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "list_monthly_paginated", "transaction", "infra", time.Since(start))
		return nil, err
	}

	for _, m := range monthlyList {
		m.LoadItems(itemsByMonthly[m.ID.String()])
	}

	r.fm.RecordRepositoryQuery(ctx, "list_monthly_paginated", "transaction", time.Since(start))
	return monthlyList, nil
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
