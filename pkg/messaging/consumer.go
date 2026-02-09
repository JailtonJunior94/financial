package messaging

import (
	"context"
)

// Consumer abstração de consumer para RabbitMQ.
//
// Esta interface define o contrato para consumidores de mensagens,
// permitindo fácil teste e manutenção do código de domínio.
//
// Exemplo de uso:
//
//	// Criar consumer (implementação específica)
//	consumer := rabbitmq.NewConsumer(client, config)
//
//	// Registrar handlers de domínio
//	consumer.RegisterHandler(budgetHandler)
//	consumer.RegisterHandler(invoiceHandler)
//
//	// Iniciar consumo
//	consumer.Start(ctx)
//
//	// Graceful shutdown
//	consumer.Shutdown(shutdownCtx)
type Consumer interface {
	// Start inicia o consumo de mensagens.
	// Deve retornar imediatamente (não-bloqueante).
	// Use goroutines para processamento de mensagens.
	Start(ctx context.Context) error

	// Shutdown para o consumer gracefully.
	// Aguarda mensagens em processamento finalizarem.
	// Deve respeitar o timeout definido no contexto.
	Shutdown(ctx context.Context) error

	// RegisterHandler registra um handler para processar mensagens.
	// Pode ser chamado múltiplas vezes para registrar vários handlers.
	// Cada handler define os topics que processa via Topics().
	RegisterHandler(handler Handler) error

	// Name retorna identificador do consumer (usado em logs).
	Name() string
}

// ConsumerConfig configuração base para consumers.
// Implementações específicas podem estender esta struct.
type ConsumerConfig struct {
	// QueueName nome da fila a ser consumida (obrigatório)
	QueueName string

	// WorkerCount número de workers concorrentes para processar mensagens
	// Default: 1 (sequencial)
	WorkerCount int

	// AutoAck se true, ACK automático (não recomendado para produção)
	// Default: false (ACK manual após processamento bem-sucedido)
	AutoAck bool

	// PrefetchCount quantas mensagens buscar antecipadamente
	// Default: 10
	PrefetchCount int
}

// DefaultConsumerConfig retorna configuração padrão do consumer.
func DefaultConsumerConfig(queueName string) *ConsumerConfig {
	return &ConsumerConfig{
		QueueName:     queueName,
		WorkerCount:   5,
		AutoAck:       false,
		PrefetchCount: 10,
	}
}
