package outbox

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent representa um evento de domínio armazenado para publicação confiável.
// Espelha EXATAMENTE a tabela outbox_events do schema.
type OutboxEvent struct {
	ID            uuid.UUID    `db:"id"`
	AggregateID   uuid.UUID    `db:"aggregate_id"`
	AggregateType string       `db:"aggregate_type"`
	EventType     string       `db:"event_type"`
	Payload       JSONBPayload `db:"payload"`
	Status        OutboxStatus `db:"status"`
	RetryCount    int          `db:"retry_count"`
	PublishedAt   *time.Time   `db:"published_at"`
	FailedAt      *time.Time   `db:"failed_at"`
	NextRetryAt   *time.Time   `db:"next_retry_at"`
	CreatedAt     time.Time    `db:"created_at"`
}

// OutboxStatus representa os possíveis status de um evento outbox.
// Valores permitidos pelo schema: 'pending', 'published', 'failed'.
type OutboxStatus string

const (
	StatusPending   OutboxStatus = "pending"
	StatusPublished OutboxStatus = "published"
	StatusFailed    OutboxStatus = "failed"
)

// String retorna a representação em string do status.
func (s OutboxStatus) String() string {
	return string(s)
}

// Valid verifica se o status é válido conforme constraint do DB.
func (s OutboxStatus) Valid() bool {
	switch s {
	case StatusPending, StatusPublished, StatusFailed:
		return true
	default:
		return false
	}
}

// JSONBPayload encapsula o payload JSONB do evento.
// Implementa sql.Scanner e driver.Valuer para compatibilidade com database/sql.
type JSONBPayload map[string]any

// Value implementa driver.Valuer para serializar para o banco.
func (j JSONBPayload) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implementa sql.Scanner para deserializar do banco.
func (j *JSONBPayload) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONBPayload: expected []byte, got %T", value)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal JSONBPayload: %w", err)
	}

	*j = payload
	return nil
}

// MaxRetryCount define o limite máximo de tentativas conforme constraint do DB.
const MaxRetryCount = 3

// retryBackoffDurations define o tempo de espera antes de cada tentativa.
// Index = retry_count após o incremento (1-based).
//
//	retry_count=1 → aguarda 30s antes da segunda tentativa
//	retry_count=2 → aguarda 2m antes da terceira tentativa
var retryBackoffDurations = [MaxRetryCount]time.Duration{
	30 * time.Second,
	2 * time.Minute,
}

// CanRetry verifica se o evento ainda pode ser retentado.
func (e *OutboxEvent) CanRetry() bool {
	return e.RetryCount < MaxRetryCount
}

// IncrementRetry incrementa o contador de tentativas e agenda o próximo retry
// com backoff exponencial. Retorna error se exceder o limite.
func (e *OutboxEvent) IncrementRetry() error {
	if !e.CanRetry() {
		return fmt.Errorf("max retry count reached: %d", MaxRetryCount)
	}
	e.RetryCount++

	// Calcular next_retry_at com base no novo retry_count.
	// O índice é (RetryCount-1) para manter o array 0-based.
	idx := e.RetryCount - 1
	if idx < len(retryBackoffDurations) {
		nextRetry := time.Now().Add(retryBackoffDurations[idx])
		e.NextRetryAt = &nextRetry
	}

	return nil
}

// MarkAsPublished marca o evento como publicado com sucesso.
func (e *OutboxEvent) MarkAsPublished() {
	now := time.Now()
	e.Status = StatusPublished
	e.PublishedAt = &now
	e.FailedAt = nil
}

// MarkAsFailed marca o evento como falho definitivamente.
func (e *OutboxEvent) MarkAsFailed() {
	now := time.Now()
	e.Status = StatusFailed
	e.FailedAt = &now
}

// MarkAsPending marca o evento como pendente (usado para retry).
func (e *OutboxEvent) MarkAsPending() {
	e.Status = StatusPending
	e.PublishedAt = nil
	e.FailedAt = nil
}

// NewOutboxEvent cria um novo evento outbox.
func NewOutboxEvent(
	aggregateID uuid.UUID,
	aggregateType string,
	eventType string,
	payload JSONBPayload,
) *OutboxEvent {
	return &OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payload,
		Status:        StatusPending,
		RetryCount:    0,
		NextRetryAt:   nil, // novos eventos são processados imediatamente
		CreatedAt:     time.Now(),
	}
}
