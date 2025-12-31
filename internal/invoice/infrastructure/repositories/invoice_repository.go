package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/database"
)

type invoiceRepository struct {
	exec database.DBExecutor
	o11y observability.Observability
}

func NewInvoiceRepository(exec database.DBExecutor, o11y observability.Observability) interfaces.InvoiceRepository {
	return &invoiceRepository{
		exec: exec,
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

	_, err := r.exec.ExecContext(
		ctx,
		query,
		invoice.ID.Value,
		invoice.UserID.Value,
		invoice.CardID.Value,
		invoice.ReferenceMonth.ToTime(),
		invoice.DueDate,
		invoice.TotalAmount.Cents(),
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

	query := `insert into invoice_items (
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
	) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, item := range items {
		_, err := r.exec.ExecContext(
			ctx,
			query,
			item.ID.Value,
			item.InvoiceID.Value,
			item.CategoryID.Value,
			item.PurchaseDate,
			item.Description,
			item.TotalAmount.Cents(),
			item.InstallmentNumber,
			item.InstallmentTotal,
			item.InstallmentAmount.Cents(),
			item.CreatedAt,
			item.UpdatedAt.Ptr(),
			item.DeletedAt.Ptr(),
		)
		if err != nil {
			return err
		}
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

	row := r.exec.QueryRowContext(ctx, query, id.Value)

	invoice, err := r.scanInvoice(row)
	if err != nil {
		if err == sql.ErrNoRows {
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

	row := r.exec.QueryRowContext(ctx, query, userID.Value, cardID.Value, referenceMonth.String())

	invoice, err := r.scanInvoice(row)
	if err != nil {
		if err == sql.ErrNoRows {
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

	rows, err := r.exec.QueryContext(ctx, query, userID.Value, referenceMonth.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*entities.Invoice
	for rows.Next() {
		invoice, err := r.scanInvoiceFromRows(rows)
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

	rows, err := r.exec.QueryContext(ctx, query, cardID.Value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*entities.Invoice
	for rows.Next() {
		invoice, err := r.scanInvoiceFromRows(rows)
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

func (r *invoiceRepository) Update(ctx context.Context, invoice *entities.Invoice) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "invoice_repository.update")
	defer span.End()

	query := `update invoices set
		total_amount = $2,
		updated_at = $3
	where id = $1`

	_, err := r.exec.ExecContext(
		ctx,
		query,
		invoice.ID.Value,
		invoice.TotalAmount.Cents(),
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

	_, err := r.exec.ExecContext(
		ctx,
		query,
		item.ID.Value,
		item.CategoryID.Value,
		item.Description,
		item.TotalAmount.Cents(),
		item.InstallmentAmount.Cents(),
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

	_, err := r.exec.ExecContext(ctx, query, itemID.Value, time.Now().UTC())
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

	rows, err := r.exec.QueryContext(ctx, query, purchaseDate, categoryID.Value, description)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	rows, err := r.exec.QueryContext(ctx, query, invoiceID.Value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r *invoiceRepository) scanInvoice(row *sql.Row) (*entities.Invoice, error) {
	var invoice entities.Invoice
	var updatedAt, deletedAt *time.Time
	var totalAmountCents int64
	var referenceDate time.Time

	err := row.Scan(
		&invoice.ID.Value,
		&invoice.UserID.Value,
		&invoice.CardID.Value,
		&referenceDate,
		&invoice.DueDate,
		&totalAmountCents,
		&invoice.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	totalAmountFloat := float64(totalAmountCents) / 100.0
	invoice.TotalAmount, _ = vos.NewMoneyFromFloat(totalAmountFloat, "BRL")
	invoice.ReferenceMonth = invoiceVos.NewReferenceMonthFromDate(referenceDate)

	if updatedAt != nil {
		invoice.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		invoice.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	return &invoice, nil
}

func (r *invoiceRepository) scanInvoiceFromRows(rows *sql.Rows) (*entities.Invoice, error) {
	var invoice entities.Invoice
	var updatedAt, deletedAt *time.Time
	var totalAmountCents int64
	var referenceDate time.Time

	err := rows.Scan(
		&invoice.ID.Value,
		&invoice.UserID.Value,
		&invoice.CardID.Value,
		&referenceDate,
		&invoice.DueDate,
		&totalAmountCents,
		&invoice.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	totalAmountFloat := float64(totalAmountCents) / 100.0
	invoice.TotalAmount, _ = vos.NewMoneyFromFloat(totalAmountFloat, "BRL")
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
	var totalAmountCents, installmentAmountCents int64

	err := rows.Scan(
		&item.ID.Value,
		&item.InvoiceID.Value,
		&item.CategoryID.Value,
		&item.PurchaseDate,
		&item.Description,
		&totalAmountCents,
		&item.InstallmentNumber,
		&item.InstallmentTotal,
		&installmentAmountCents,
		&item.CreatedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	totalAmountFloat := float64(totalAmountCents) / 100.0
	installmentAmountFloat := float64(installmentAmountCents) / 100.0

	item.TotalAmount, _ = vos.NewMoneyFromFloat(totalAmountFloat, "BRL")
	item.InstallmentAmount, _ = vos.NewMoneyFromFloat(installmentAmountFloat, "BRL")

	if updatedAt != nil {
		item.UpdatedAt = vos.NewNullableTime(*updatedAt)
	}
	if deletedAt != nil {
		item.DeletedAt = vos.NewNullableTime(*deletedAt)
	}

	return &item, nil
}
