package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"

	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// Builder facilita criação de consumers RabbitMQ.
// Encapsula declaração de topologia (exchange, queue, bindings).
//
// Exemplo de uso:
//
//	builder := rabbitmq.NewBuilder(client, o11y)
//	consumer, _ := builder.BuildConsumer(ctx, config)
type Builder struct {
	client *rabbitmq.Client
	o11y   observability.Observability
}

// NewBuilder cria builder de consumers.
func NewBuilder(client *rabbitmq.Client, o11y observability.Observability) *Builder {
	return &Builder{
		client: client,
		o11y:   o11y,
	}
}

// BuildConsumer cria consumer configurado com topologia declarada.
// Declara exchange, queue e bindings automaticamente.
func (b *Builder) BuildConsumer(ctx context.Context, config *ConsumerConfig) (messaging.Consumer, error) {
	if config == nil {
		return nil, fmt.Errorf("consumer config cannot be nil")
	}

	// Declara exchange se especificado
	if config.Exchange != "" {
		if err := b.client.DeclareExchange(
			ctx,
			config.Exchange,
			"topic", // Tipo topic para suportar routing patterns
			config.Durable,
			false, // Não auto-delete exchange
			nil,
		); err != nil {
			b.o11y.Logger().Error(ctx,
				"failed to declare exchange",
				observability.String("exchange", config.Exchange),
				observability.Error(err),
			)
			return nil, fmt.Errorf("failed to declare exchange %s: %w", config.Exchange, err)
		}

		b.o11y.Logger().Debug(ctx,
			"exchange declared",
			observability.String("exchange", config.Exchange),
		)
	}

	// Declara queue
	if _, err := b.client.DeclareQueue(
		ctx,
		config.QueueName,
		config.Durable,
		config.AutoDelete,
		false, // Não exclusive
		nil,
	); err != nil {
		b.o11y.Logger().Error(ctx,
			"failed to declare queue",
			observability.String("queue", config.QueueName),
			observability.Error(err),
		)
		return nil, fmt.Errorf("failed to declare queue %s: %w", config.QueueName, err)
	}

	b.o11y.Logger().Debug(ctx,
		"queue declared",
		observability.String("queue", config.QueueName),
	)

	// Bind routing keys ao exchange
	if config.Exchange != "" && len(config.RoutingKeys) > 0 {
		for _, routingKey := range config.RoutingKeys {
			if err := b.client.BindQueue(
				ctx,
				config.QueueName,
				routingKey,
				config.Exchange,
				nil,
			); err != nil {
				b.o11y.Logger().Error(ctx,
					"failed to bind queue",
					observability.String("queue", config.QueueName),
					observability.String("routing_key", routingKey),
					observability.String("exchange", config.Exchange),
					observability.Error(err),
				)
				return nil, fmt.Errorf("failed to bind queue %s to %s with key %s: %w",
					config.QueueName, config.Exchange, routingKey, err)
			}

			b.o11y.Logger().Debug(ctx,
				"queue bound to exchange",
				observability.String("queue", config.QueueName),
				observability.String("routing_key", routingKey),
				observability.String("exchange", config.Exchange),
			)
		}
	}

	// Cria consumer
	consumer, err := NewConsumer(b.client, config, b.o11y)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	b.o11y.Logger().Info(ctx,
		"consumer created successfully",
		observability.String("queue", config.QueueName),
		observability.String("exchange", config.Exchange),
		observability.Int("routing_keys_count", len(config.RoutingKeys)),
	)

	return consumer, nil
}
