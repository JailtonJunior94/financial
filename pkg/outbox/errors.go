package outbox

import "errors"

var (
	// ErrEventNotFound indica que o evento outbox não foi encontrado.
	ErrEventNotFound = errors.New("outbox event not found")

	// ErrInvalidStatus indica que o status fornecido é inválido.
	ErrInvalidStatus = errors.New("invalid outbox status")

	// ErrMaxRetriesReached indica que o evento atingiu o limite de tentativas.
	ErrMaxRetriesReached = errors.New("max retry count reached")

	// ErrNoEventsToProcess indica que não há eventos pendentes para processar.
	ErrNoEventsToProcess = errors.New("no pending events to process")

	// ErrPublishFailed indica falha na publicação do evento.
	ErrPublishFailed = errors.New("failed to publish event")

	// ErrInvalidPayload indica que o payload do evento é inválido.
	ErrInvalidPayload = errors.New("invalid event payload")
)
