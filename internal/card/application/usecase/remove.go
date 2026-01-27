package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	RemoveCardUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewRemoveCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) RemoveCardUseCase {
	return &removeCardUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *removeCardUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_card_usecase.execute")
	defer span.End()

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
			panic(err)
		}
	}()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
		u.metrics.RecordError(ctx, metrics.OperationDelete, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing user id",
			observability.String("user_id", userID),
			observability.Error(err),
		)

		return err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
		u.metrics.RecordError(ctx, metrics.OperationDelete, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing card id",
			observability.String("card_id", id),
			observability.Error(err),
		)

		return err
	}

	card, err := u.repository.FindByID(ctx, user, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
		u.metrics.RecordError(ctx, metrics.OperationDelete, metrics.ClassifyError(err))

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
		return err
	}

	if card == nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
		u.metrics.RecordError(ctx, metrics.OperationDelete, metrics.ErrorTypeNotFound)

		span.AddEvent(
			"card not found",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Error(
			ctx,
			"card not found",
			observability.Error(customErrors.ErrCardNotFound),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return customErrors.ErrCardNotFound
	}

	if err := u.repository.Update(ctx, card.Delete()); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration)
		u.metrics.RecordError(ctx, metrics.OperationDelete, metrics.ClassifyError(err))

		span.AddEvent(
			"error deleting card in repository",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(
			ctx,
			"error deleting card in repository",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationDelete, duration)
	u.metrics.DecActiveCards(ctx)

	return nil
}
