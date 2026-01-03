package kafka

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// Builder facilita criação de consumers Kafka.
type Builder struct {
	config *ConsumerConfig
	o11y   observability.Observability
}

// NewBuilder cria novo builder.
func NewBuilder(config *ConsumerConfig, o11y observability.Observability) *Builder {
	return &Builder{
		config: config,
		o11y:   o11y,
	}
}

// BuildConsumer cria consumer Kafka.
//
// TODO: Implementar criação de tópicos, consumer groups, etc.
//
// Exemplo de implementação futura:
//
//	func (b *Builder) BuildConsumer(ctx context.Context) (messaging.Consumer, error) {
//	    // Criar admin client para verificar/criar tópico
//	    admin := kafka.NewAdmin(b.config.Brokers)
//	    defer admin.Close()
//
//	    // Verificar se tópico existe
//	    topics, _ := admin.ListTopics()
//	    if !contains(topics, b.config.Topic) {
//	        // Criar tópico
//	        admin.CreateTopic(b.config.Topic, ...)
//	    }
//
//	    // Criar consumer
//	    return NewConsumer(b.config, b.o11y)
//	}
func (b *Builder) BuildConsumer(ctx context.Context) (messaging.Consumer, error) {
	b.o11y.Logger().Warn(ctx,
		"building kafka consumer (not implemented)",
		observability.String("topic", b.config.Topic),
		observability.String("group_id", b.config.GroupID),
	)

	// TODO: Implementar setup de infraestrutura Kafka
	consumer, err := NewConsumer(b.config, b.o11y)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return consumer, nil
}
