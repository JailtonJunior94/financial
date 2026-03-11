package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/google/uuid"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type transactionRepository struct {
	db   database.DBTX
	o11y observability.Observability
	tm   *metrics.TransactionMetrics
}

// NewTransactionRepository creates a new TransactionRepository.
func NewTransactionRepository(db database.DBTX, o11y observability.Observability, tm *metrics.TransactionMetrics) interfaces.TransactionRepository {
	return &transactionRepository{db: db, o11y: o11y, tm: tm}
}

func (r *transactionRepository) Save(ctx context.Context, tx database.DBTX, t *entities.Transaction) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.save")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
		observability.String("user_id", t.UserID.String()),
	)

	query := `
		INSERT INTO transactions (
			id, user_id, category_id, subcategory_id, card_id,
			invoice_id, installment_group_id, description, amount,
			payment_method, transaction_date, installment_number, installment_total,
			status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "save", "transaction", "infra", time.Since(start))
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
			r.o11y.Logger().Error(ctx, "Save: failed to close stmt",
				observability.Error(closeErr),
			)
		}
	}()

	_, err = stmt.ExecContext(ctx,
		t.ID.Value,
		t.UserID.Value,
		t.CategoryID.Value,
		optionalUUID(t.SubcategoryID),
		optionalUUID(t.CardID),
		optionalUUID(t.InvoiceID),
		optionalUUID(t.InstallmentGroupID),
		t.Description,
		t.Amount.Float(),
		t.PaymentMethod.String(),
		t.TransactionDate,
		t.InstallmentNumber,
		t.InstallmentTotal,
		t.Status.String(),
		t.CreatedAt,
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "save", "transaction", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "save", "transaction", time.Since(start))
	return nil
}

func (r *transactionRepository) SaveAll(ctx context.Context, tx database.DBTX, ts []*entities.Transaction) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.save_all")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "save_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)

	for _, t := range ts {
		if err := r.Save(ctx, tx, t); err != nil {
			span.RecordError(err)
			r.tm.RecordRepositoryFailure(ctx, "save_all", "transaction", "infra", time.Since(start))
			return err
		}
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "save_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "save_all", "transaction", time.Since(start))
	return nil
}

func (r *transactionRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Transaction, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_by_id")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)

	query := `
		SELECT id, user_id, category_id, subcategory_id, card_id,
		       invoice_id, installment_group_id, description, amount,
		       payment_method, transaction_date, installment_number, installment_total,
		       status, created_at, updated_at, deleted_at
		FROM transactions
		WHERE id = $1 AND deleted_at IS NULL`

	row := r.db.QueryRowContext(ctx, query, id.Value)
	t, err := r.scanTransaction(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.tm.RecordRepositoryQuery(ctx, "find_by_id", "transaction", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "find_by_id", "transaction", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "find_by_id", "transaction", time.Since(start))
	return t, nil
}

func (r *transactionRepository) FindByInstallmentGroup(ctx context.Context, groupID vos.UUID) ([]*entities.Transaction, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.find_by_group")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_group"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)

	query := `
		SELECT id, user_id, category_id, subcategory_id, card_id,
		       invoice_id, installment_group_id, description, amount,
		       payment_method, transaction_date, installment_number, installment_total,
		       status, created_at, updated_at, deleted_at
		FROM transactions
		WHERE installment_group_id = $1 AND deleted_at IS NULL
		ORDER BY installment_number ASC`

	rows, err := r.db.QueryContext(ctx, query, groupID.Value)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_group"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "find_by_group", "transaction", "infra", time.Since(start))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
			r.o11y.Logger().Error(ctx, "FindByInstallmentGroup: failed to close rows",
				observability.Error(closeErr),
			)
		}
	}()

	var transactions []*entities.Transaction
	for rows.Next() {
		t, err := r.scanTransaction(rows)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "find_by_group"),
				observability.String("layer", "repository"),
				observability.String("entity", "transaction"),
				observability.Error(err),
			)
			r.tm.RecordRepositoryFailure(ctx, "find_by_group", "transaction", "infra", time.Since(start))
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.tm.RecordRepositoryFailure(ctx, "find_by_group", "transaction", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_group"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "find_by_group", "transaction", time.Since(start))
	return transactions, nil
}

func (r *transactionRepository) Update(ctx context.Context, tx database.DBTX, t *entities.Transaction) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.update")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
		observability.String("user_id", t.UserID.String()),
	)

	query := `
		UPDATE transactions SET
			description = $2,
			amount = $3,
			category_id = $4,
			subcategory_id = $5,
			status = $6,
			updated_at = NOW()
		WHERE id = $1`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "update", "transaction", "infra", time.Since(start))
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
			r.o11y.Logger().Error(ctx, "Update: failed to close stmt",
				observability.Error(closeErr),
			)
		}
	}()

	_, err = stmt.ExecContext(ctx,
		t.ID.Value,
		t.Description,
		t.Amount.Float(),
		t.CategoryID.Value,
		optionalUUID(t.SubcategoryID),
		t.Status.String(),
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "update", "transaction", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "update", "transaction", time.Since(start))
	return nil
}

func (r *transactionRepository) UpdateAll(ctx context.Context, tx database.DBTX, ts []*entities.Transaction) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.update_all")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "update_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)

	for _, t := range ts {
		if err := r.Update(ctx, tx, t); err != nil {
			span.RecordError(err)
			r.tm.RecordRepositoryFailure(ctx, "update_all", "transaction", "infra", time.Since(start))
			return err
		}
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "update_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "update_all", "transaction", time.Since(start))
	return nil
}

func (r *transactionRepository) ListPaginated(ctx context.Context, params interfaces.ListParams) ([]*entities.Transaction, string, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "transaction_repository.list_paginated")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
		observability.String("user_id", params.UserID.String()),
	)

	conditions := []string{"user_id = $1", "deleted_at IS NULL", "status = 'active'"}
	args := []any{params.UserID.Value}
	argIdx := 2

	if params.PaymentMethod != "" {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIdx))
		args = append(args, params.PaymentMethod)
		argIdx++
	}
	if params.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, params.CategoryID)
		argIdx++
	}
	if params.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argIdx))
		args = append(args, *params.StartDate)
		argIdx++
	}
	if params.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argIdx))
		args = append(args, *params.EndDate)
		argIdx++
	}

	cursor, err := pagination.DecodeCursor(params.Cursor)
	if err == nil {
		if txDate, ok := cursor.GetString("transaction_date"); ok {
			if cursorID, ok := cursor.GetString("id"); ok && txDate != "" && cursorID != "" {
				conditions = append(conditions,
					fmt.Sprintf("(transaction_date, id) < ($%d, $%d)", argIdx, argIdx+1))
				args = append(args, txDate, cursorID)
				argIdx += 2
			}
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, category_id, subcategory_id, card_id,
		       invoice_id, installment_group_id, description, amount,
		       payment_method, transaction_date, installment_number, installment_total,
		       status, created_at, updated_at, deleted_at
		FROM transactions
		WHERE %s
		ORDER BY transaction_date DESC, id DESC
		LIMIT $%d`,
		strings.Join(conditions, " AND "), argIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list_paginated"),
			observability.String("layer", "repository"),
			observability.String("entity", "transaction"),
			observability.Error(err),
		)
		r.tm.RecordRepositoryFailure(ctx, "list_paginated", "transaction", "infra", time.Since(start))
		return nil, "", err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
			r.o11y.Logger().Error(ctx, "ListPaginated: failed to close rows",
				observability.Error(closeErr),
			)
		}
	}()

	transactions := make([]*entities.Transaction, 0, limit)
	for rows.Next() {
		t, err := r.scanTransaction(rows)
		if err != nil {
			span.RecordError(err)
			r.tm.RecordRepositoryFailure(ctx, "list_paginated", "transaction", "infra", time.Since(start))
			return nil, "", err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.tm.RecordRepositoryFailure(ctx, "list_paginated", "transaction", "infra", time.Since(start))
		return nil, "", err
	}

	var nextCursor string
	if len(transactions) > limit {
		transactions = transactions[:limit]
		last := transactions[len(transactions)-1]
		c := pagination.Cursor{
			Fields: map[string]any{
				"transaction_date": last.TransactionDate.Format("2006-01-02"),
				"id":               last.ID.String(),
			},
		}
		encoded, encErr := pagination.EncodeCursor(c)
		if encErr == nil {
			nextCursor = encoded
		}
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "transaction"),
	)
	r.tm.RecordRepositoryQuery(ctx, "list_paginated", "transaction", time.Since(start))
	return transactions, nextCursor, nil
}

type transactionScanner interface {
	Scan(dest ...any) error
}

func (r *transactionRepository) scanTransaction(s transactionScanner) (*entities.Transaction, error) {
	var t entities.Transaction
	var subcategoryID, cardID, invoiceID, installmentGroupID *uuid.UUID
	var installmentNumber, installmentTotal *int
	var updatedAt, deletedAt *time.Time
	var amountStr string
	var paymentMethodStr, statusStr string

	err := s.Scan(
		&t.ID.Value,
		&t.UserID.Value,
		&t.CategoryID.Value,
		&subcategoryID,
		&cardID,
		&invoiceID,
		&installmentGroupID,
		&t.Description,
		&amountStr,
		&paymentMethodStr,
		&t.TransactionDate,
		&installmentNumber,
		&installmentTotal,
		&statusStr,
		&t.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	amount, err := vos.NewMoneyFromString(amountStr, vos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}
	t.Amount = amount

	pm, err := transactionVos.NewPaymentMethod(paymentMethodStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment_method: %w", err)
	}
	t.PaymentMethod = pm

	status, err := transactionVos.NewTransactionStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}
	t.Status = status

	if subcategoryID != nil {
		uid := vos.UUID{Value: *subcategoryID}
		t.SubcategoryID = &uid
	}
	if cardID != nil {
		uid := vos.UUID{Value: *cardID}
		t.CardID = &uid
	}
	if invoiceID != nil {
		uid := vos.UUID{Value: *invoiceID}
		t.InvoiceID = &uid
	}
	if installmentGroupID != nil {
		uid := vos.UUID{Value: *installmentGroupID}
		t.InstallmentGroupID = &uid
	}

	t.InstallmentNumber = installmentNumber
	t.InstallmentTotal = installmentTotal
	t.UpdatedAt = updatedAt
	t.DeletedAt = deletedAt

	return &t, nil
}

func optionalUUID(u *vos.UUID) *uuid.UUID {
	if u == nil {
		return nil
	}
	return &u.Value
}
