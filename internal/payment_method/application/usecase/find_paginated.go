package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	// FindPaymentMethodPaginatedUseCase lista payment methods com paginação cursor-based.
	FindPaymentMethodPaginatedUseCase interface {
		Execute(ctx context.Context, input FindPaymentMethodPaginatedInput) (*FindPaymentMethodPaginatedOutput, error)
	}

	// FindPaymentMethodPaginatedInput representa a entrada do use case.
	FindPaymentMethodPaginatedInput struct {
		Limit  int
		Cursor string
		Code   string // Opcional: filtra por código
	}

	// FindPaymentMethodPaginatedOutput representa a saída do use case.
	FindPaymentMethodPaginatedOutput struct {
		PaymentMethods []*dtos.PaymentMethodOutput
		NextCursor     *string
	}

	findPaymentMethodPaginatedUseCase struct {
		o11y       observability.Observability
		repository interfaces.PaymentMethodRepository
	}
)

// NewFindPaymentMethodPaginatedUseCase cria uma nova instância do use case.
func NewFindPaymentMethodPaginatedUseCase(
	o11y observability.Observability,
	repository interfaces.PaymentMethodRepository,
) FindPaymentMethodPaginatedUseCase {
	return &findPaymentMethodPaginatedUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

// Execute executa o use case de listagem paginada de payment methods.
func (u *findPaymentMethodPaginatedUseCase) Execute(
	ctx context.Context,
	input FindPaymentMethodPaginatedInput,
) (*FindPaymentMethodPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_payment_method_paginated_usecase.execute")
	defer span.End()

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}

	// List payment methods (paginado)
	paymentMethods, err := u.repository.ListPaginated(ctx, interfaces.ListPaymentMethodsParams{
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
		Code:   input.Code,
	})
	if err != nil {
		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(paymentMethods) > input.Limit
	if hasNext {
		paymentMethods = paymentMethods[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(paymentMethods) > 0 {
		lastPM := paymentMethods[len(paymentMethods)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"name": lastPM.Name.String(),
				"id":   lastPM.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	paymentMethodsOutput := linq.Map(paymentMethods, func(pm *entities.PaymentMethod) *dtos.PaymentMethodOutput {
		output := &dtos.PaymentMethodOutput{
			ID:          pm.ID.String(),
			Name:        pm.Name.String(),
			Code:        pm.Code.String(),
			Description: pm.Description.String(),
			CreatedAt:   pm.CreatedAt.ValueOr(time.Time{}),
		}
		if !pm.UpdatedAt.ValueOr(time.Time{}).IsZero() {
			output.UpdatedAt = pm.UpdatedAt.ValueOr(time.Time{})
		}
		return output
	})

	return &FindPaymentMethodPaginatedOutput{
		PaymentMethods: paymentMethodsOutput,
		NextCursor:     nextCursor,
	}, nil
}
