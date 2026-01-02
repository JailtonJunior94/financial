package outbox

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/google/uuid"
)

type Service interface {
	SaveEvent(ctx context.Context, tx database.DBTX, event *OutboxEvent) error
	SaveDomainEvent(
		ctx context.Context,
		tx database.DBTX,
		aggregateID uuid.UUID,
		aggregateType string,
		eventType string,
		payload JSONBPayload,
	) error
}

type service struct {
	repository Repository
	o11y       observability.Observability
}

// NewService cria uma nova instância do service.
func NewService(repository Repository, o11y observability.Observability) Service {
	return &service{
		o11y:       o11y,
		repository: repository,
	}
}

// SaveEvent persiste um evento outbox dentro de uma transação.
func (s *service) SaveEvent(ctx context.Context, tx database.DBTX, event *OutboxEvent) error {
	ctx, span := s.o11y.Tracer().Start(ctx, "outbox.service.save_event")
	defer span.End()

	// Validar status
	if !event.Status.Valid() {
		s.o11y.Logger().Error(ctx, "invalid outbox status",
			observability.String("status", event.Status.String()),
		)
		return ErrInvalidStatus
	}

	// Validar payload
	if len(event.Payload) == 0 {
		s.o11y.Logger().Error(ctx, "invalid outbox payload")
		return ErrInvalidPayload
	}

	// Persistir evento
	if err := s.repository.Save(ctx, event); err != nil {
		s.o11y.Logger().Error(ctx, "failed to save outbox event",
			observability.Error(err),
			observability.String("aggregate_type", event.AggregateType),
			observability.String("event_type", event.EventType),
		)
		return fmt.Errorf("save outbox event: %w", err)
	}

	s.o11y.Logger().Info(ctx, "outbox event saved",
		observability.String("event_id", event.ID.String()),
		observability.String("aggregate_id", event.AggregateID.String()),
		observability.String("aggregate_type", event.AggregateType),
		observability.String("event_type", event.EventType),
	)

	return nil
}

// SaveDomainEvent cria e salva um evento de domínio.
func (s *service) SaveDomainEvent(
	ctx context.Context,
	tx database.DBTX,
	aggregateID uuid.UUID,
	aggregateType string,
	eventType string,
	payload JSONBPayload,
) error {
	event := NewOutboxEvent(aggregateID, aggregateType, eventType, payload)
	return s.SaveEvent(ctx, tx, event)
}
