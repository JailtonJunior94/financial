package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCardUseCase interface {
		Execute(ctx context.Context, userID string) ([]*dtos.CardOutput, error)
	}

	findCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewFindCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) FindCardUseCase {
	return &findCardUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *findCardUseCase) Execute(ctx context.Context, userID string) ([]*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_card_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing user id",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	cards, err := u.repository.List(ctx, user)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFind, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error listing cards from repository",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	cardsOutput := linq.Map(cards, func(card *entities.Card) *dtos.CardOutput {
		output := &dtos.CardOutput{
			ID:                card.ID.String(),
			Name:              card.Name.String(),
			DueDay:            card.DueDay.Int(),
			ClosingOffsetDays: card.ClosingOffsetDays.Int(),
			CreatedAt:         card.CreatedAt.ValueOr(time.Time{}),
		}
		if !card.UpdatedAt.ValueOr(time.Time{}).IsZero() {
			output.UpdatedAt = card.UpdatedAt.ValueOr(time.Time{})
		}
		return output
	})

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationFind, duration)

	return cardsOutput, nil
}
