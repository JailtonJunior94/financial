package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type paymentMethodRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewPaymentMethodRepository(db database.DBTX, o11y observability.Observability) interfaces.PaymentMethodRepository {
	return &paymentMethodRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *paymentMethodRepository) List(ctx context.Context) ([]*entities.PaymentMethod, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.list")
	defer span.End()

	query := `select
				id,
				name,
				code,
				description,
				created_at,
				updated_at,
				deleted_at
			from
				payment_methods
			where
				deleted_at is null
			order by
				name;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		span.RecordError(err)

		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)

		}
	}()

	var paymentMethods []*entities.PaymentMethod
	for rows.Next() {
		var pm entities.PaymentMethod
		err := rows.Scan(
			&pm.ID.Value,
			&pm.Name.Value,
			&pm.Code.Value,
			&pm.Description.Value,
			&pm.CreatedAt,
			&pm.UpdatedAt,
			&pm.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)

			return nil, err
		}
		paymentMethods = append(paymentMethods, &pm)
	}
	return paymentMethods, nil
}

func (r *paymentMethodRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.PaymentMethod, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.find_by_id")
	defer span.End()

	query := `select
				id,
				name,
				code,
				description,
				created_at,
				updated_at,
				deleted_at
			from
				payment_methods
			where
				deleted_at is null
				and id = $1;`

	var pm entities.PaymentMethod
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&pm.ID.Value,
		&pm.Name.Value,
		&pm.Code.Value,
		&pm.Description.Value,
		&pm.CreatedAt,
		&pm.UpdatedAt,
		&pm.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		span.RecordError(err)

		return nil, err
	}

	return &pm, nil
}

func (r *paymentMethodRepository) FindByCode(ctx context.Context, code string) (*entities.PaymentMethod, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.find_by_code")
	defer span.End()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))

	query := `select
				id,
				name,
				code,
				description,
				created_at,
				updated_at,
				deleted_at
			from
				payment_methods
			where
				deleted_at is null
				and code = $1;`

	var pm entities.PaymentMethod
	err := r.db.QueryRowContext(ctx, query, normalizedCode).Scan(
		&pm.ID.Value,
		&pm.Name.Value,
		&pm.Code.Value,
		&pm.Description.Value,
		&pm.CreatedAt,
		&pm.UpdatedAt,
		&pm.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		span.RecordError(err)

		return nil, err
	}

	return &pm, nil
}

func (r *paymentMethodRepository) Save(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.save")
	defer span.End()

	query := `insert into
				payment_methods (
					id,
					name,
					code,
					description,
					created_at,
					updated_at,
					deleted_at
				)
				values
					($1, $2, $3, $4, $5, $6, $7)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddEvent(
			"error preparing insert payment method",
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		paymentMethod.ID.Value,
		paymentMethod.Name.Value,
		paymentMethod.Code.Value,
		paymentMethod.Description.Value,
		paymentMethod.CreatedAt.Ptr(),
		paymentMethod.UpdatedAt.Ptr(),
		paymentMethod.DeletedAt.Ptr(),
	)
	if err != nil {
		span.AddEvent(
			"error inserting payment method",
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	return nil
}

func (r *paymentMethodRepository) Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.update")
	defer span.End()

	query := `update
				payment_methods
			set
				name = $1,
				description = $2,
				updated_at = $3,
				deleted_at = $4
			where
				id = $5`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddEvent(
			"error preparing update payment method",
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		paymentMethod.Name.Value,
		paymentMethod.Description.Value,
		paymentMethod.UpdatedAt.Ptr(),
		paymentMethod.DeletedAt.Ptr(),
		paymentMethod.ID.Value,
	)
	if err != nil {
		span.AddEvent(
			"error updating payment method",
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	return nil
}

// ListPaginated lista payment methods com paginação cursor-based.
func (r *paymentMethodRepository) ListPaginated(
	ctx context.Context,
	params interfaces.ListPaymentMethodsParams,
) ([]*entities.PaymentMethod, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "payment_method_repository.list_paginated")
	defer span.End()

	// Build WHERE clause
	whereClause := "deleted_at is null"
	args := []interface{}{}
	argIndex := 1

	// Filter by code if provided
	if params.Code != "" {
		whereClause += fmt.Sprintf(" AND code = $%d", argIndex)
		args = append(args, params.Code)
		argIndex++
	}

	// Add cursor conditions
	cursorName, hasName := params.Cursor.GetString("name")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasName && hasID && cursorName != "" && cursorID != "" {
		whereClause += fmt.Sprintf(` AND (
			name > $%d
			OR (name = $%d AND id > $%d)
		)`, argIndex, argIndex, argIndex+1)
		args = append(args, cursorName, cursorID)
		argIndex += 2
	}

	query := fmt.Sprintf(`
		SELECT
			id, name, code, description,
			created_at, updated_at, deleted_at
		FROM payment_methods
		WHERE %s
		ORDER BY name ASC, id ASC
		LIMIT $%d`, whereClause, argIndex)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var paymentMethods []*entities.PaymentMethod
	for rows.Next() {
		var pm entities.PaymentMethod
		err := rows.Scan(
			&pm.ID.Value,
			&pm.Name.Value,
			&pm.Code.Value,
			&pm.Description.Value,
			&pm.CreatedAt,
			&pm.UpdatedAt,
			&pm.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		paymentMethods = append(paymentMethods, &pm)
	}

	return paymentMethods, rows.Err()
}
