package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	// FindCardPaginatedUseCase lista cards de um usuário com paginação cursor-based.
	FindCardPaginatedUseCase interface {
		Execute(ctx context.Context, input FindCardPaginatedInput) (*FindCardPaginatedOutput, error)
	}

	// FindCardPaginatedInput representa a entrada do use case.
	FindCardPaginatedInput struct {
		UserID string
		Limit  int
		Cursor string
	}

	// FindCardPaginatedOutput representa a saída do use case.
	FindCardPaginatedOutput struct {
		Cards      []*dtos.CardOutput
		NextCursor *string
	}

	findCardPaginatedUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

// NewFindCardPaginatedUseCase cria uma nova instância do use case.
func NewFindCardPaginatedUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) FindCardPaginatedUseCase {
	return &findCardPaginatedUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

// Execute executa o use case de listagem paginada de cards.
func (u *findCardPaginatedUseCase) Execute(
	ctx context.Context,
	input FindCardPaginatedInput,
) (*FindCardPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_card_paginated_usecase.execute")
	defer span.End()

	start := time.Now()

	// Parse user ID
	userID, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing user id",
			observability.String("user_id", input.UserID),
			observability.Error(err),
		)

		return nil, err
	}

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error decoding cursor",
			observability.String("cursor", input.Cursor),
			observability.Error(err),
		)

		return nil, err
	}

	// List cards (paginado)
	cards, err := u.repository.ListPaginated(ctx, interfaces.ListCardsParams{
		UserID: userID,
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
	})
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error listing cards from repository",
			observability.String("user_id", input.UserID),
			observability.Error(err),
		)

		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(cards) > input.Limit
	if hasNext {
		cards = cards[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(cards) > 0 {
		lastCard := cards[len(cards)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"name": lastCard.Name.String(),
				"id":   lastCard.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			duration := time.Since(start)
			u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

			span.AddEvent(
				"error encoding cursor",
				observability.Error(err),
			)

			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	output := make([]*dtos.CardOutput, len(cards))
	for i, card := range cards {
		output[i] = &dtos.CardOutput{
			ID:                card.ID.String(),
			Name:              card.Name.String(),
			DueDay:            card.DueDay.Int(),
			ClosingOffsetDays: card.ClosingOffsetDays.Int(),
			CreatedAt:         card.CreatedAt.ValueOr(time.Time{}),
		}
		if !card.UpdatedAt.ValueOr(time.Time{}).IsZero() {
			output[i].UpdatedAt = card.UpdatedAt.ValueOr(time.Time{})
		}
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationFind, duration)

	return &FindCardPaginatedOutput{
		Cards:      output,
		NextCursor: nextCursor,
	}, nil
}
