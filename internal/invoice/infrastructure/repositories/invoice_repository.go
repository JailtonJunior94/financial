package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

const (
	// currencyScale é usado para converter centavos para valor decimal (100 centavos = 1.00).
	currencyScale = 100.0
)

// scanner é uma interface comum para *sql.Row e *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

type invoiceRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewInvoiceRepository(db database.DBTX, o11y observability.Observability) interfaces.InvoiceRepository {
	return &invoiceRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *invoiceRepository) Insert(ctx context.Context, invoice *entities.Invoice) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.insert")
	defer span.End()

	query := `insert into invoices (
		id,
		user_id,
		card_id,
		reference_month,
		due_date,
		total_amount,
		created_at,
		updated_at,
		deleted_at
	) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		invoice.ID.Value,
		invoice.UserID.Value,
		invoice.CardID.Value,
		invoice.ReferenceMonth.ToTime(),
		invoice.DueDate,
		float64(invoice.TotalAmount.Cents())/currencyScale,
		invoice.CreatedAt,
		invoice.UpdatedAt.Ptr(),
		invoice.DeletedAt.Ptr(),
	)

	return err
}

func (r *invoiceRepository) InsertItems(ctx context.Context, items []*entities.InvoiceItem) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.insert_items")
	defer span.End()

	if len(items) == 0 {
		return nil
	}

	const numColumns = 12
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
			item.InvoiceID.Value,
			item.CategoryID.Value,
			item.PurchaseDate,
			item.Description,
			float64(item.TotalAmount.Cents())/currencyScale,
			item.InstallmentNumber,
			item.InstallmentTotal,
			float64(item.InstallmentAmount.Cents())/currencyScale,
			item.CreatedAt,
			item.UpdatedAt.Ptr(),
			item.DeletedAt.Ptr(),
		)
	}

	query := fmt.Sprintf(`insert into invoice_items (
		id,
		invoice_id,
		category_id,
		purchase_date,
		description,
		total_amount,
		installment_number,
		installment_total,
		installment_amount,
		created_at,
		updated_at,
		deleted_at
	) values %s`, strings.Join(valueStrings, ", "))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *invoiceRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.find_by_id")
	defer span.End()

	query := `select
		id,
		user_id,
		card_id,
		reference_month,
		due_date,
		total_amount,
		created_at,
		updated_at,
		deleted_at
	from invoices
	where id = $1 and deleted_at is null`

	row := r.db.QueryRowContext(ctx, query, id.Value)

	invoice, err := r.scanInvoice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Carregar itens
	items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return invoice, nil
}

func (r *invoiceRepository) FindByUserAndCardAndMonth(
	ctx context.Context,
	userID vos.UUID,
	cardID vos.UUID,
	referenceMonth invoiceVos.ReferenceMonth,
) (*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.find_by_user_card_month")
	defer span.End()

	query := `select
		id,
		user_id,
		card_id,
		reference_month,
		due_date,
		total_amount,
		created_at,
		updated_at,
		deleted_at
	from invoices
	where user_id = $1
	  and card_id = $2
	  and to_char(reference_month, 'YYYY-MM') = $3
	  and deleted_at is null`

	row := r.db.QueryRowContext(ctx, query, userID.Value, cardID.Value, referenceMonth.String())

	invoice, err := r.scanInvoice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Carregar itens
	items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return invoice, nil
}

func (r *invoiceRepository) FindByUserAndMonth(
	ctx context.Context,
	userID vos.UUID,
	referenceMonth invoiceVos.ReferenceMonth,
) ([]*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.find_by_user_month")
	defer span.End()

	query := `select
		id,
		user_id,
		card_id,
		reference_month,
		due_date,
		total_amount,
		created_at,
		updated_at,
		deleted_at
	from invoices
	where user_id = $1
	  and to_char(reference_month, 'YYYY-MM') = $2
	  and deleted_at is null
	order by due_date`

	rows, err := r.db.QueryContext(ctx, query, userID.Value, referenceMonth.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var invoices []*entities.Invoice
	for rows.Next() {
		invoice, err := r.scanInvoice(rows)
		if err != nil {
			return nil, err
		}

		// Carregar itens
		items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
		if err != nil {
			return nil, err
		}
		invoice.Items = items

		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}

// ListByUserAndMonthPaginated busca faturas de um usuário em um mês com paginação cursor-based.
func (r *invoiceRepository) ListByUserAndMonthPaginated(
	ctx context.Context,
	params interfaces.ListInvoicesByMonthParams,
) ([]*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.list_by_user_month_paginated")
	defer span.End()

	// Build WHERE clause with cursor
	whereClause := `user_id = $1
		AND to_char(reference_month, 'YYYY-MM') = $2
		AND deleted_at IS NULL`
	args := []interface{}{params.UserID.Value, params.ReferenceMonth.String()}

	cursorDueDate, hasDueDate := params.Cursor.GetString("due_date")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasDueDate && hasID && cursorDueDate != "" && cursorID != "" {
		// Keyset pagination: WHERE (due_date, id) > (cursor_due_date, cursor_id)
		whereClause += ` AND (
			due_date > $3
			OR (due_date = $3 AND id > $4)
		)`
		args = append(args, cursorDueDate, cursorID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			user_id,
			card_id,
			reference_month,
			due_date,
			total_amount,
			created_at,
			updated_at,
			deleted_at
		FROM invoices
		WHERE %s
		ORDER BY due_date ASC, id ASC
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	invoices := make([]*entities.Invoice, 0)
	for rows.Next() {
		invoice, err := r.scanInvoice(rows)
		if err != nil {
			return nil, err
		}

		// Carregar itens
		items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
		if err != nil {
			return nil, err
		}
		invoice.Items = items

		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}

func (r *invoiceRepository) FindByCard(ctx context.Context, cardID vos.UUID) ([]*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.find_by_card")
	defer span.End()

	query := `select
		id,
		user_id,
		card_id,
		reference_month,
		due_date,
		total_amount,
		created_at,
		updated_at,
		deleted_at
	from invoices
	where card_id = $1 and deleted_at is null
	order by reference_month desc`

	rows, err := r.db.QueryContext(ctx, query, cardID.Value)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var invoices []*entities.Invoice
	for rows.Next() {
		invoice, err := r.scanInvoice(rows)
		if err != nil {
			return nil, err
		}

		// Carregar itens
		items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
		if err != nil {
			return nil, err
		}
		invoice.Items = items

		invoices = append(invoices, invoice)
	}

	return invoices, rows.Err()
}

// ListByCard busca faturas de um cartão com cursor-based pagination.
// Ordena por reference_month DESC, id DESC (mais recentes primeiro).
func (r *invoiceRepository) ListByCard(ctx context.Context, params interfaces.ListInvoicesByCardParams) ([]*entities.Invoice, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.list_by_card")
	defer span.End()

	// Base query com ORDER BY determinístico
	query := `
		SELECT
			id,
			user_id,
			card_id,
			reference_month,
			due_date,
			total_amount,
			created_at,
			updated_at,
			deleted_at
		FROM invoices
		WHERE
			card_id = $1
			AND deleted_at IS NULL`

	args := []interface{}{params.CardID}
	argIndex := 2

	// Adicionar condição de cursor (keyset pagination)
	// Para ORDER BY DESC, usamos < em vez de >
	if refMonth, ok := params.Cursor.GetString("reference_month"); ok {
		if id, ok := params.Cursor.GetString("id"); ok {
			// WHERE (reference_month, id) < (cursor_month, cursor_id)
			query += fmt.Sprintf(` AND (reference_month, id) < ($%d, $%d)`, argIndex, argIndex+1)
			args = append(args, refMonth, id)
			argIndex += 2
		}
	}

	// ORDER BY determinístico (reference_month DESC, id DESC)
	query += ` ORDER BY reference_month DESC, id DESC`

	// LIMIT
	query += fmt.Sprintf(` LIMIT $%d`, argIndex)
	args = append(args, params.Limit)

	// Executar query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	// Scanear resultados
	var invoices []*entities.Invoice
	for rows.Next() {
		invoice, err := r.scanInvoice(rows)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		// Carregar itens da fatura
		items, err := r.findItemsByInvoiceID(ctx, invoice.ID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		invoice.Items = items

		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Garantir que retorna array vazio em vez de nil
	if invoices == nil {
		invoices = []*entities.Invoice{}
	}

	return invoices, nil
}

func (r *invoiceRepository) Update(ctx context.Context, invoice *entities.Invoice) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.update")
	defer span.End()

	query := `update invoices set
		total_amount = $2,
		updated_at = $3
	where id = $1`

	_, err := r.db.ExecContext(
		ctx,
		query,
		invoice.ID.Value,
		float64(invoice.TotalAmount.Cents())/currencyScale,
		time.Now().UTC(),
	)

	return err
}

func (r *invoiceRepository) UpdateItem(ctx context.Context, item *entities.InvoiceItem) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.update_item")
	defer span.End()

	query := `update invoice_items set
		category_id = $2,
		description = $3,
		total_amount = $4,
		installment_amount = $5,
		updated_at = $6
	where id = $1`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID.Value,
		item.CategoryID.Value,
		item.Description,
		float64(item.TotalAmount.Cents())/currencyScale,
		float64(item.InstallmentAmount.Cents())/currencyScale,
		time.Now().UTC(),
	)

	return err
}

func (r *invoiceRepository) DeleteItem(ctx context.Context, itemID vos.UUID) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.delete_item")
	defer span.End()

	query := `update invoice_items set
		deleted_at = $2
	where id = $1`

	_, err := r.db.ExecContext(ctx, query, itemID.Value, time.Now().UTC())
	return err
}

func (r *invoiceRepository) FindItemsByPurchaseOrigin(
	ctx context.Context,
	purchaseDate string,
	categoryID vos.UUID,
	description string,
) ([]*entities.InvoiceItem, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.find_items_by_purchase_origin")
	defer span.End()

	query := `select
		id,
		invoice_id,
		category_id,
		purchase_date,
		description,
		total_amount,
		installment_number,
		installment_total,
		installment_amount,
		created_at,
		updated_at,
		deleted_at
	from invoice_items
	where purchase_date = $1
	  and category_id = $2
	  and description = $3
	  and deleted_at is null
	order by installment_number`

	rows, err := r.db.QueryContext(ctx, query, purchaseDate, categoryID.Value, description)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.InvoiceItem
	for rows.Next() {
		item, err := r.scanInvoiceItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *invoiceRepository) findItemsByInvoiceID(ctx context.Context, invoiceID vos.UUID) ([]*entities.InvoiceItem, error) {
	query := `select
		id,
		invoice_id,
		category_id,
		purchase_date,
		description,
		total_amount,
		installment_number,
		installment_total,
		installment_amount,
		created_at,
		updated_at,
		deleted_at
	from invoice_items
	where invoice_id = $1 and deleted_at is null
	order by purchase_date, installment_number`

	rows, err := r.db.QueryContext(ctx, query, invoiceID.Value)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.InvoiceItem
	for rows.Next() {
		item, err := r.scanInvoiceItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// scanInvoice é um helper unificado que funciona com *sql.Row e *sql.Rows.
func (r *invoiceRepository) scanInvoice(s scanner) (*entities.Invoice, error) {
	var invoice entities.Invoice
	var updatedAt, deletedAt *time.Time
	var totalAmount string
	var referenceDate time.Time

	err := s.Scan(
		&invoice.ID.Value,
		&invoice.UserID.Value,
		&invoice.CardID.Value,
		&referenceDate,
		&invoice.DueDate,
		&totalAmount,
		&invoice.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	invoice.TotalAmount, err = vos.NewMoneyFromString(totalAmount, "BRL")
	if err != nil {
		return nil, fmt.Errorf("failed to create Money from total_amount: %w", err)
	}

	invoice.ReferenceMonth = invoiceVos.NewReferenceMonthFromDate(referenceDate)

	if updatedAt != nil {
		invoice.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		invoice.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	return &invoice, nil
}

func (r *invoiceRepository) scanInvoiceItemFromRows(rows *sql.Rows) (*entities.InvoiceItem, error) {
	var item entities.InvoiceItem
	var updatedAt, deletedAt *time.Time
	var totalAmount, installmentAmount string

	err := rows.Scan(
		&item.ID.Value,
		&item.InvoiceID.Value,
		&item.CategoryID.Value,
		&item.PurchaseDate,
		&item.Description,
		&totalAmount,
		&item.InstallmentNumber,
		&item.InstallmentTotal,
		&installmentAmount,
		&item.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	item.TotalAmount, err = vos.NewMoneyFromString(totalAmount, "BRL")
	if err != nil {
		return nil, fmt.Errorf("failed to create Money from total_amount: %w", err)
	}

	item.InstallmentAmount, err = vos.NewMoneyFromString(installmentAmount, "BRL")
	if err != nil {
		return nil, fmt.Errorf("failed to create Money from installment_amount: %w", err)
	}

	if updatedAt != nil {
		item.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		item.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	return &item, nil
}
