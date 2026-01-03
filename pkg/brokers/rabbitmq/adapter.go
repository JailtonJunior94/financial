package rabbitmq

import (
	"context"

	"github.com/jailtonjunior94/financial/pkg/lifecycle"
	"github.com/jailtonjunior94/financial/pkg/messaging"
)

// ConsumerService adapta messaging.Consumer para lifecycle.Service.
// Permite gerenciar consumers RabbitMQ usando lifecycle.Group.
//
// Exemplo de uso:
//
//	consumer, _ := rabbitmq.NewConsumer(client, config, o11y)
//	service := rabbitmq.NewConsumerService(consumer)
//
//	group := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())
//	group.Register(service)
//	group.Start(ctx)
type ConsumerService struct {
	consumer messaging.Consumer
}

// NewConsumerService cria adapter de Consumer para Service.
// Segue o padrão adapter (igual a pkg/outbox/jobs.go).
func NewConsumerService(consumer messaging.Consumer) lifecycle.Service {
	return &ConsumerService{consumer: consumer}
}

// Start inicia o consumer (implementa lifecycle.Service).
func (s *ConsumerService) Start(ctx context.Context) error {
	return s.consumer.Start(ctx)
}

// Shutdown para o consumer (implementa lifecycle.Service).
func (s *ConsumerService) Shutdown(ctx context.Context) error {
	return s.consumer.Shutdown(ctx)
}

// Name retorna nome do service (implementa lifecycle.Service).
func (s *ConsumerService) Name() string {
	return s.consumer.Name()
}
