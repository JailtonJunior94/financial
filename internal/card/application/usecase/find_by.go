package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCardByUseCase interface {
		Execute(ctx context.Context, userID, id string) (*dtos.CardOutput, error)
	}

	findCardByUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewFindCardByUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) FindCardByUseCase {
	return &findCardByUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *findCardByUseCase) Execute(ctx context.Context, userID, id string) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_card_by_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing user id",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return nil, err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing card id",
			observability.String("card_id", id),
			observability.Error(err),
		)

		return nil, err
	}

	card, err := u.repository.FindByID(ctx, user, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error finding card by id",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(ctx, "error finding card by id",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, err
	}

	if card == nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationFindBy, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"card not found",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Error(ctx, "card not found",
			observability.Error(customErrors.ErrCardNotFound),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return nil, customErrors.ErrCardNotFound
	}

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

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationFindBy, duration)

	return output, nil
}
