package sqs

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/pkg/messaging"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// ConsumerConfig configuração específica do SQS consumer.
type ConsumerConfig struct {
	// QueueURL URL completa da fila SQS
	// Ex: https://sqs.us-east-1.amazonaws.com/123456789/financial-queue
	QueueURL string

	// Region região AWS (ex: us-east-1)
	Region string

	// WorkerCount número de workers concorrentes para processar mensagens
	// Default: 5
	WorkerCount int

	// MaxMessages número máximo de mensagens a buscar por poll
	// Default: 10 (máximo permitido pelo SQS)
	MaxMessages int32

	// WaitTimeSeconds long polling wait time
	// Default: 20 (recomendado para reduzir custos)
	WaitTimeSeconds int32

	// VisibilityTimeout tempo em segundos que mensagem fica invisível após recebida
	// Default: 30
	VisibilityTimeout int32
}

// DefaultConsumerConfig retorna configuração padrão.
func DefaultConsumerConfig(queueURL, region string) *ConsumerConfig {
	return &ConsumerConfig{
		QueueURL:          queueURL,
		Region:            region,
		WorkerCount:       5,
		MaxMessages:       10,
		WaitTimeSeconds:   20,
		VisibilityTimeout: 30,
	}
}

// Consumer implementação SQS do messaging.Consumer.
//
// TODO: Implementar usando aws-sdk-go-v2
//
// Exemplo de implementação futura:
//
//	import (
//	    "github.com/aws/aws-sdk-go-v2/config"
//	    "github.com/aws/aws-sdk-go-v2/service/sqs"
//	)
//
//	type Consumer struct {
//	    client  *sqs.Client
//	    config  *ConsumerConfig
//	    handler messaging.Handler
//	    o11y    observability.Observability
//	}
//
//	func (c *Consumer) Start(ctx context.Context) error {
//	    go func() {
//	        for {
//	            // Receive messages
//	            result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
//	                QueueUrl:            &c.config.QueueURL,
//	                MaxNumberOfMessages: c.config.MaxMessages,
//	                WaitTimeSeconds:     c.config.WaitTimeSeconds,
//	            })
//
//	            for _, msg := range result.Messages {
//	                genericMsg := &messaging.Message{
//	                    ID:      *msg.MessageId,
//	                    Payload: []byte(*msg.Body),
//	                    Headers: convertSQSAttributes(msg.MessageAttributes),
//	                }
//
//	                if err := c.handler.Handle(ctx, genericMsg); err != nil {
//	                    // Log error, retry
//	                } else {
//	                    // Delete message (ACK)
//	                    c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
//	                        QueueUrl:      &c.config.QueueURL,
//	                        ReceiptHandle: msg.ReceiptHandle,
//	                    })
//	                }
//	            }
//	        }
//	    }()
//	    return nil
//	}
type Consumer struct {
	config  *ConsumerConfig
	handler messaging.Handler
	o11y    observability.Observability
}

// NewConsumer cria consumer SQS.
func NewConsumer(
	config *ConsumerConfig,
	o11y observability.Observability,
) (*Consumer, error) {
	if config == nil {
		return nil, fmt.Errorf("consumer config cannot be nil")
	}

	if config.QueueURL == "" {
		return nil, fmt.Errorf("queue URL is required")
	}

	if config.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	return &Consumer{
		config: config,
		o11y:   o11y,
	}, nil
}

// RegisterHandler registra handler.
func (c *Consumer) RegisterHandler(handler messaging.Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	c.handler = handler

	c.o11y.Logger().Info(context.Background(),
		"handler registered",
		observability.String("queue_url", c.config.QueueURL),
		observability.Int("topics_count", len(handler.Topics())),
	)

	return nil
}

// Start inicia consumer.
func (c *Consumer) Start(ctx context.Context) error {
	c.o11y.Logger().Warn(ctx,
		"sqs consumer not implemented yet",
		observability.String("queue_url", c.config.QueueURL),
	)

	// TODO: Implementar consumo real com aws-sdk-go-v2
	return fmt.Errorf("sqs consumer not implemented yet")
}

// Shutdown para consumer.
func (c *Consumer) Shutdown(ctx context.Context) error {
	c.o11y.Logger().Info(ctx,
		"shutting down sqs consumer",
		observability.String("queue_url", c.config.QueueURL),
	)

	// TODO: Implementar shutdown (aguardar workers, etc)
	return nil
}

// Name retorna identificador do consumer.
func (c *Consumer) Name() string {
	return fmt.Sprintf("sqs-consumer-%s", c.config.QueueURL)
}
