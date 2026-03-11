package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/factories"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateCardUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error)
	}

	createCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
		metrics    *metrics.CardMetrics
	}
)

func NewCreateCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
	metrics *metrics.CardMetrics,
) CreateCardUseCase {
	return &createCardUseCase{
		o11y:       o11y,
		repository: repository,
		metrics:    metrics,
	}
}

func (u *createCardUseCase) Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_card_usecase.execute")
	defer span.End()

	start := time.Now()

	dueDay := 0
	if input.DueDay != nil {
		dueDay = *input.DueDay
	}

	closingOffsetDays := 0
	if input.ClosingOffsetDays != nil {
		closingOffsetDays = *input.ClosingOffsetDays
	}

	card, err := factories.CreateCard(factories.CreateCardParams{
		UserID:            userID,
		Name:              input.Name,
		Type:              input.Type,
		Flag:              input.Flag,
		LastFourDigits:    input.LastFourDigits,
		DueDay:            dueDay,
		ClosingOffsetDays: closingOffsetDays,
	})
	if err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration, metrics.ClassifyError(err))

		span.RecordError(err)

		return nil, err
	}

	if err := u.repository.Save(ctx, card); err != nil {
		duration := time.Since(start)
		u.metrics.RecordOperationFailure(ctx, metrics.OperationCreate, duration, metrics.ClassifyError(err))

		span.RecordError(err)

		return nil, err
	}

	duration := time.Since(start)
	u.metrics.RecordOperation(ctx, metrics.OperationCreate, duration)
	u.metrics.IncActiveCards(ctx)

	output := &dtos.CardOutput{
		ID:             card.ID.String(),
		Name:           card.Name.String(),
		Type:           card.Type.Value,
		Flag:           card.Flag.Value,
		LastFourDigits: card.LastFourDigits.Value,
		CreatedAt:      card.CreatedAt.ValueOr(time.Time{}),
	}
	if card.Type.IsCredit() {
		dueDay := card.DueDay.Int()
		output.DueDay = &dueDay
		offset := card.ClosingOffsetDays.Int()
		output.ClosingOffsetDays = &offset
	}

	return output, nil
}
