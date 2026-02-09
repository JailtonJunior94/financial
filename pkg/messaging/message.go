package messaging

import (
	"time"

	"github.com/google/uuid"
)

// Message representa uma mensagem do RabbitMQ.
// Abstração que facilita o processamento de mensagens sem acoplar ao broker.
//
// Exemplo de uso:
//
//	msg := messaging.NewMessage("user.created", payload)
//	msg.WithHeaders(map[string]any{
//	    "source": "user-service",
//	}).WithCorrelationID(correlationID)
type Message struct {
	// ID identificador único da mensagem
	ID string

	// Topic routing key ou tópico (ex: "transaction.created", "invoice.added")
	Topic string

	// Payload dados da mensagem (JSON serializado)
	Payload []byte

	// Headers metadados adicionais
	Headers map[string]any

	// Timestamp quando a mensagem foi criada
	Timestamp time.Time

	// DeliveryAttempt número de tentativas de entrega (1 = primeira vez)
	DeliveryAttempt int

	// CorrelationID para rastreamento distribuído
	CorrelationID string
}

// NewMessage cria uma nova mensagem.
// topic: routing key ou tópico (ex: "transaction.created")
// payload: dados da mensagem (geralmente JSON serializado)
func NewMessage(topic string, payload []byte) *Message {
	return &Message{
		ID:              uuid.New().String(),
		Topic:           topic,
		Payload:         payload,
		Headers:         make(map[string]any),
		Timestamp:       time.Now(),
		DeliveryAttempt: 1,
		CorrelationID:   uuid.New().String(),
	}
}

// WithHeaders adiciona headers à mensagem.
// Retorna a própria mensagem para permitir chaining.
func (m *Message) WithHeaders(headers map[string]any) *Message {
	m.Headers = headers
	return m
}

// WithCorrelationID define correlation ID para rastreamento distribuído.
// Retorna a própria mensagem para permitir chaining.
func (m *Message) WithCorrelationID(id string) *Message {
	m.CorrelationID = id
	return m
}

// WithID define ID customizado da mensagem.
// Retorna a própria mensagem para permitir chaining.
func (m *Message) WithID(id string) *Message {
	m.ID = id
	return m
}

// WithTimestamp define timestamp customizado.
// Retorna a própria mensagem para permitir chaining.
func (m *Message) WithTimestamp(timestamp time.Time) *Message {
	m.Timestamp = timestamp
	return m
}

// WithDeliveryAttempt define número de tentativas de entrega.
// Retorna a própria mensagem para permitir chaining.
func (m *Message) WithDeliveryAttempt(attempt int) *Message {
	m.DeliveryAttempt = attempt
	return m
}
