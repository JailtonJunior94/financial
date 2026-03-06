package usecase

import (
	"context"
	"time"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
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
		o11y           observability.Observability
		repository     interfaces.CardRepository
		invoiceChecker interfaces.InvoiceChecker
		metrics        *metrics.CardMetrics
	}
)

func NewRemoveCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	invoiceChecker interfaces.InvoiceChecker,
	metrics *metrics.CardMetrics,
) RemoveCardUseCase {
	return &removeCardUseCase{
		o11y:           o11y,
		repository:     repository,
		invoiceChecker: invoiceChecker,
		metrics:        metrics,
	}
}

func (u *removeCardUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_card_usecase.execute")
	defer span.End()

	start := time.Now()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ClassifyError(err))

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
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error parsing card id",
			observability.String("card_id", id),
			observability.Error(err),
		)

		return err
	}

	card, err := u.repository.FindByIDOnly(ctx, cardID)
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error finding card by id",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "RemoveCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		return err
	}

	if card == nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ErrorTypeNotFound)

		span.AddEvent(
			"card not found",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Warn(ctx, "card not found",
			observability.String("operation", "RemoveCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return customErrors.ErrCardNotFound
	}

	if card.UserID.String() != user.String() {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, "authorization")

		span.AddEvent(
			"card ownership mismatch",
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		u.o11y.Logger().Warn(ctx, "card ownership mismatch",
			observability.String("operation", "RemoveCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
		)
		return customErrors.ErrForbidden
	}

	if card.Type.IsCredit() {
		hasOpen, err := u.invoiceChecker.HasOpenInvoices(ctx, card.ID)
		if err != nil {
			duration := time.Since(start)
			u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ClassifyError(err))

			span.AddEvent(
				"error checking open invoices",
				observability.String("user_id", userID),
				observability.String("card_id", id),
				observability.Error(err),
			)
			u.o11y.Logger().Error(ctx, "invoice_check_failed",
				observability.String("operation", "RemoveCard"),
				observability.String("layer", "usecase"),
				observability.String("entity", "card"),
				observability.String("user_id", userID),
				observability.String("card_id", id),
				observability.Error(err),
			)
			return err
		}

		if hasOpen {
			duration := time.Since(start)
			u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, "business")

			u.o11y.Logger().Warn(ctx, "card_has_open_invoices",
				observability.String("operation", "RemoveCard"),
				observability.String("layer", "usecase"),
				observability.String("entity", "card"),
				observability.String("user_id", userID),
				observability.String("card_id", id),
			)
			return domain.ErrCardHasOpenInvoices
		}
	}

	if err := u.repository.Update(ctx, card.Delete()); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationDelete, duration, metrics.ClassifyError(err))

		span.AddEvent(
			"error deleting card in repository",
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		u.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "RemoveCard"),
			observability.String("layer", "usecase"),
			observability.String("entity", "card"),
			observability.String("user_id", userID),
			observability.String("card_id", id),
			observability.Error(err),
		)
		return err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationDelete, duration)
	u.metrics.DecActiveCards(ctx)

	return nil
}
