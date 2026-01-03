package rabbitmq

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// SimpleConsumer implementação simplificada de consumer.
// NOTA: Esta é uma implementação simplificada que delega o gerenciamento
// real para o código existente em cmd/consumer. O devkit-go já possui
// todo o necessário, esta camada serve apenas para manter a interface
// messaging.Consumer consistente.
//
// Para uso real, veja cmd/consumer/consumers.go que usa devkit-go diretamente.
type SimpleConsumer struct {
	config   *ConsumerConfig
	handlers []messaging.Handler
	o11y     observability.Observability
}

// NewSimpleConsumer cria consumer simplificado.
// Este consumer mantém apenas os handlers registrados e delega
// a execução para o código em cmd/consumer.
func NewSimpleConsumer(
	config *ConsumerConfig,
	o11y observability.Observability,
) *SimpleConsumer {
	return &SimpleConsumer{
		config:   config,
		handlers: make([]messaging.Handler, 0),
		o11y:     o11y,
	}
}

// RegisterHandler registra handler.
func (c *SimpleConsumer) RegisterHandler(handler messaging.Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	c.handlers = append(c.handlers, handler)

	c.o11y.Logger().Info(context.Background(),
		"handler registered",
		observability.String("queue", c.config.QueueName),
		observability.Int("topics_count", len(handler.Topics())),
	)

	return nil
}

// Start não implementado (delegado para cmd/consumer).
func (c *SimpleConsumer) Start(ctx context.Context) error {
	c.o11y.Logger().Info(ctx,
		"simple consumer start called",
		observability.String("queue", c.config.QueueName),
	)
	// A implementação real está em cmd/consumer/consumers.go
	return nil
}

// Shutdown não implementado (delegado para cmd/consumer).
func (c *SimpleConsumer) Shutdown(ctx context.Context) error {
	c.o11y.Logger().Info(ctx,
		"simple consumer shutdown called",
		observability.String("queue", c.config.QueueName),
	)
	return nil
}

// Name retorna identificador.
func (c *SimpleConsumer) Name() string {
	return fmt.Sprintf("simple-consumer-%s", c.config.QueueName)
}

// Handlers retorna handlers registrados (para uso em cmd/consumer).
func (c *SimpleConsumer) Handlers() []messaging.Handler {
	return c.handlers
}
