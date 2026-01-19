package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// DispatcherConfig configura o comportamento do dispatcher.
type DispatcherConfig struct {
	// BatchSize quantidade de eventos a processar por execução
	BatchSize int

	// MaxRetries número máximo de tentativas (deve ser <= MaxRetryCount)
	MaxRetries int

	// RetryBackoff intervalo base para backoff exponencial
	RetryBackoff time.Duration

	// Exchange RabbitMQ exchange para publicação
	Exchange string
}

// DefaultDispatcherConfig retorna configuração padrão.
func DefaultDispatcherConfig(exchange string) *DispatcherConfig {
	return &DispatcherConfig{
		BatchSize:    100,
		MaxRetries:   MaxRetryCount,
		RetryBackoff: 5 * time.Second,
		Exchange:     exchange,
	}
}

// Dispatcher processa eventos pendentes e publica no message broker.
type Dispatcher interface {
	// Dispatch busca eventos pendentes e publica no broker.
	// Retorna quantidade de eventos processados.
	Dispatch(ctx context.Context) (int, error)
}

type dispatcher struct {
	uow       uow.UnitOfWork
	publisher *rabbitmq.Publisher
	config    *DispatcherConfig
	o11y      observability.Observability
}

// NewDispatcher cria uma nova instância do dispatcher.
func NewDispatcher(
	uow uow.UnitOfWork,
	rabbitClient *rabbitmq.Client,
	config *DispatcherConfig,
	o11y observability.Observability,
) Dispatcher {
	return &dispatcher{
		uow:       uow,
		o11y:      o11y,
		config:    config,
		publisher: rabbitmq.NewPublisher(rabbitClient),
	}
}

// Dispatch processa batch de eventos pendentes.
func (d *dispatcher) Dispatch(ctx context.Context) (int, error) {
	ctx, span := d.o11y.Tracer().Start(ctx, "outbox.dispatcher.dispatch")
	defer span.End()

	processed := 0

	err := d.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		outboxRepository := NewRepository(tx, d.o11y)

		events, err := outboxRepository.FindPendingBatch(ctx, d.config.BatchSize)
		if err != nil {
			return fmt.Errorf("find pending batch: %w", err)
		}

		if len(events) == 0 {
			d.o11y.Logger().Info(ctx, "no pending events to dispatch")
			return nil
		}

		// Processar cada evento
		for _, event := range events {
			if err := d.processEvent(ctx, outboxRepository, event); err != nil {
				d.o11y.Logger().Error(ctx, "failed to process event",
					observability.Error(err),
					observability.String("event_id", event.ID.String()),
				)
				continue
			}
			processed++
		}

		return nil
	})

	if err != nil {
		d.o11y.Logger().Error(ctx, "dispatch failed", observability.Error(err))
		return processed, fmt.Errorf("dispatch: %w", err)
	}

	d.o11y.Logger().Info(ctx, "dispatch completed",
		observability.Int("processed", processed),
	)

	return processed, nil
}

// processEvent processa um único evento.
func (d *dispatcher) processEvent(ctx context.Context, repository Repository, event *OutboxEvent) error {
	ctx, span := d.o11y.Tracer().Start(ctx, "outbox.dispatcher.process_event")
	defer span.End()

	// Serializar payload
	payloadBytes, err := event.Payload.Value()
	if err != nil {
		return d.handlePublishFailure(ctx, repository, event, fmt.Errorf("serialize payload: %w", err))
	}

	payload, ok := payloadBytes.([]byte)
	if !ok {
		return d.handlePublishFailure(ctx, repository, event, fmt.Errorf("invalid payload type"))
	}

	// Publicar no RabbitMQ
	routingKey := d.buildRoutingKey(event)
	publishStart := time.Now()

	err = d.publisher.Publish(ctx, d.config.Exchange, routingKey, payload,
		rabbitmq.WithContentType("application/json"),
		rabbitmq.WithDeliveryMode(2), // Persistent
		rabbitmq.WithMessageID(event.ID.String()),
		rabbitmq.WithHeaders(map[string]any{
			"aggregate_id":   event.AggregateID.String(),
			"aggregate_type": event.AggregateType,
			"event_type":     event.EventType,
		}),
	)

	publishDuration := time.Since(publishStart)

	if err != nil {
		return d.handlePublishFailure(ctx, repository, event, err)
	}

	// Marcar como publicado
	event.MarkAsPublished()
	if err := repository.UpdateStatus(ctx, event); err != nil {
		// Reverter estado em memória se persistência falhar
		event.MarkAsPending()
		return fmt.Errorf("update status to published: %w", err)
	}

	d.o11y.Logger().Info(ctx, "event published successfully",
		observability.String("event_id", event.ID.String()),
		observability.String("aggregate_type", event.AggregateType),
		observability.String("event_type", event.EventType),
		observability.String("routing_key", routingKey),
		observability.Int64("publish_duration_ms", publishDuration.Milliseconds()),
	)

	return nil
}

// handlePublishFailure trata falha na publicação com retry ou falha definitiva.
func (d *dispatcher) handlePublishFailure(ctx context.Context, repository Repository, event *OutboxEvent, err error) error {
	if event.CanRetry() {
		// Incrementar retry
		if retryErr := event.IncrementRetry(); retryErr != nil {
			return retryErr
		}

		// Manter como pending para próxima tentativa
		event.MarkAsPending()

		d.o11y.Logger().Warn(ctx, "event publish failed, will retry",
			observability.Error(err),
			observability.String("event_id", event.ID.String()),
			observability.Int("retry_count", event.RetryCount),
		)
	} else {
		// Esgotou retries, marcar como failed
		event.MarkAsFailed()

		d.o11y.Logger().Error(ctx, "event publish failed permanently",
			observability.Error(err),
			observability.String("event_id", event.ID.String()),
			observability.Int("retry_count", event.RetryCount),
		)
	}

	// Atualizar status no banco
	if updateErr := repository.UpdateStatus(ctx, event); updateErr != nil {
		return fmt.Errorf("update status after failure: %w", updateErr)
	}

	return nil
}

// buildRoutingKey constrói a routing key para o evento.
// Formato: {aggregate_type}.{event_type}.
func (d *dispatcher) buildRoutingKey(event *OutboxEvent) string {
	return fmt.Sprintf("%s.%s", event.AggregateType, event.EventType)
}
