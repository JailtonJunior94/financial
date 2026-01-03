package kafka

import (
	"context"

	"github.com/jailtonjunior94/financial/pkg/lifecycle"
	"github.com/jailtonjunior94/financial/pkg/messaging"
)

// ConsumerService adapta messaging.Consumer para lifecycle.Service.
// Padrão: Adapter pattern (igual a pkg/outbox/jobs.go e pkg/brokers/rabbitmq/adapter.go).
type ConsumerService struct {
	consumer messaging.Consumer
}

// NewConsumerService cria adapter de consumer para service.
func NewConsumerService(consumer messaging.Consumer) lifecycle.Service {
	return &ConsumerService{consumer: consumer}
}

// Start inicia consumer.
func (s *ConsumerService) Start(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

// Shutdown para consumer gracefully.
func (s *ConsumerService) Shutdown(ctx context.Context) error {
	return s.consumer.Shutdown(ctx)
}

// Name retorna identificador do service.
func (s *ConsumerService) Name() string {
	return s.consumer.Name()
}
