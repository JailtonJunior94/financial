package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"

	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// ConsumerConfig configuração específica do RabbitMQ consumer.
type ConsumerConfig struct {
	// QueueName nome da fila a ser consumida (obrigatório)
	QueueName string

	// Exchange nome do exchange
	Exchange string

	// RoutingKeys keys para bind da fila ao exchange
	// Ex: ["transaction.created", "invoice.added"]
	// Suporta patterns: ["transaction.*", "invoice.*"]
	RoutingKeys []string

	// WorkerCount número de workers concorrentes
	// Default: 5
	WorkerCount int

	// PrefetchCount quantas mensagens buscar antecipadamente
	// Default: 10
	PrefetchCount int

	// Durable fila durável (sobrevive a restart do broker)
	// Default: true (recomendado para produção)
	Durable bool

	// AutoDelete fila auto-delete quando não houver consumers
	// Default: false (recomendado para produção)
	AutoDelete bool
}

// DefaultConsumerConfig retorna configuração padrão.
func DefaultConsumerConfig(queueName string) *ConsumerConfig {
	return &ConsumerConfig{
		QueueName:     queueName,
		WorkerCount:   5,
		PrefetchCount: 10,
		Durable:       true,
		AutoDelete:    false,
	}
}

// Consumer thin adapter sobre devkit-go/messaging/rabbitmq.Consumer.
// Implementa messaging.Consumer delegando operações para devkit-go.
//
// Exemplo de uso:
//
//	client, _ := rabbitmq.New(o11y, rabbitmq.WithCloudConnection(url))
//	config := rabbitmq.DefaultConsumerConfig("budget.updates")
//	config.Exchange = "financial.events"
//	config.RoutingKeys = []string{"transaction.created"}
//
//	consumer, _ := rabbitmq.NewConsumer(client, config, o11y)
//	consumer.RegisterHandler(budgetHandler)
//	consumer.Start(ctx)
type Consumer struct {
	consumer *rabbitmq.Consumer // devkit-go consumer (faz o trabalho pesado)
	config   *ConsumerConfig
	client   *rabbitmq.Client
	o11y     observability.Observability
}

// NewConsumer cria consumer RabbitMQ usando devkit-go.
// Delega 100% das operações para devkit-go (retry, DLQ, worker pool, etc).
func NewConsumer(
	client *rabbitmq.Client,
	config *ConsumerConfig,
	o11y observability.Observability,
) (*Consumer, error) {
	if config == nil {
		return nil, fmt.Errorf("consumer config cannot be nil")
	}

	if config.QueueName == "" {
		return nil, fmt.Errorf("queue name is required")
	}

	if client == nil {
		return nil, fmt.Errorf("rabbitmq client cannot be nil")
	}

	// Cria consumer devkit-go com opções
	consumer := rabbitmq.NewConsumer(
		client,
		rabbitmq.WithQueue(config.QueueName),
		rabbitmq.WithPrefetchCount(config.PrefetchCount),
		rabbitmq.WithWorkerPool(config.WorkerCount),
		rabbitmq.WithAutoAck(false), // Sempre manual ACK (mais seguro)
	)

	return &Consumer{
		consumer: consumer,
		config:   config,
		client:   client,
		o11y:     o11y,
	}, nil
}

// RegisterHandler registra handler convertendo para formato devkit-go.
// Registra o handler para todos os topics retornados por handler.Topics().
func (c *Consumer) RegisterHandler(handler messaging.Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	topics := handler.Topics()
	if len(topics) == 0 {
		c.o11y.Logger().Warn(context.Background(),
			"handler registered without topics",
		)
		return nil
	}

	// Converte handler genérico para handler devkit-go
	for _, topic := range topics {
		devHandler := c.createDevkitHandler(handler, topic)
		c.consumer.RegisterHandler(topic, devHandler)

		c.o11y.Logger().Debug(context.Background(),
			"handler registered for topic",
			observability.String("queue", c.config.QueueName),
			observability.String("topic", topic),
		)
	}

	c.o11y.Logger().Info(context.Background(),
		"handler registered",
		observability.String("queue", c.config.QueueName),
		observability.Int("topics_count", len(topics)),
	)

	return nil
}

// createDevkitHandler cria handler devkit-go a partir de handler genérico.
// Converte rabbitmq.Message → messaging.Message.
func (c *Consumer) createDevkitHandler(handler messaging.Handler, topic string) func(context.Context, rabbitmq.Message) error {
	return func(ctx context.Context, msg rabbitmq.Message) error {
		// Converte rabbitmq.Message → messaging.Message
		genericMsg := &messaging.Message{
			Topic:   topic,
			Payload: msg.Body,
			Headers: c.convertHeaders(msg.Headers),
		}

		// Processa mensagem com handler genérico
		return handler.Handle(ctx, genericMsg)
	}
}

// convertHeaders converte headers do RabbitMQ para formato genérico.
func (c *Consumer) convertHeaders(headers map[string]interface{}) map[string]interface{} {
	if headers == nil {
		return make(map[string]interface{})
	}
	return headers
}

// Start inicia consumer (delega para devkit-go).
// Não-bloqueante: retorna imediatamente.
func (c *Consumer) Start(ctx context.Context) error {
	c.o11y.Logger().Info(ctx,
		"starting rabbitmq consumer",
		observability.String("queue", c.config.QueueName),
		observability.Int("worker_count", c.config.WorkerCount),
	)

	// Delega para devkit-go
	return c.consumer.Consume(ctx)
}

// Shutdown para consumer (delega para devkit-go).
// Aguarda mensagens em processamento finalizarem.
func (c *Consumer) Shutdown(ctx context.Context) error {
	c.o11y.Logger().Info(ctx,
		"shutting down rabbitmq consumer",
		observability.String("queue", c.config.QueueName),
	)

	// Delega para devkit-go
	return c.consumer.Close()
}

// Name retorna identificador do consumer.
func (c *Consumer) Name() string {
	return fmt.Sprintf("rabbitmq-consumer-%s", c.config.QueueName)
}
