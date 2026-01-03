package sqs

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// Builder facilita criação de consumers SQS.
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

// BuildConsumer cria consumer SQS.
//
// TODO: Implementar criação/verificação de fila, DLQ, etc.
//
// Exemplo de implementação futura:
//
//	func (b *Builder) BuildConsumer(ctx context.Context) (messaging.Consumer, error) {
//	    // Criar AWS config
//	    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(b.config.Region))
//
//	    // Criar SQS client
//	    client := sqs.NewFromConfig(cfg)
//
//	    // Verificar se fila existe
//	    _, err = client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
//	        QueueUrl: &b.config.QueueURL,
//	    })
//	    if err != nil {
//	        return nil, fmt.Errorf("queue not found: %w", err)
//	    }
//
//	    // Criar consumer
//	    return NewConsumer(client, b.config, b.o11y)
//	}
func (b *Builder) BuildConsumer(ctx context.Context) (messaging.Consumer, error) {
	b.o11y.Logger().Warn(ctx,
		"building sqs consumer (not implemented)",
		observability.String("queue_url", b.config.QueueURL),
		observability.String("region", b.config.Region),
	)

	// TODO: Implementar setup de infraestrutura SQS
	consumer, err := NewConsumer(b.config, b.o11y)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs consumer: %w", err)
	}

	return consumer, nil
}
