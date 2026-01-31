package metrics

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// CardMetrics agrupa todas as métricas do módulo de cartões (OpenTelemetry)
type CardMetrics struct {
	o11y observability.Observability

	// Instrumentos OpenTelemetry
	operationsTotal   observability.Counter
	operationDuration observability.Histogram
	activeCardsTotal  observability.UpDownCounter
}

// NewCardMetrics inicializa métricas OpenTelemetry para o módulo de cartões
func NewCardMetrics(o11y observability.Observability) *CardMetrics {
	metrics := &CardMetrics{
		o11y: o11y,
	}

	// Counter: Operações executadas (inclui status e error_type quando aplicável)
	metrics.operationsTotal = o11y.Metrics().Counter(
		"financial.card.operations.total",
		"Total number of card operations by type, status, and error_type",
		"operation",
	)

	// Histogram: Latência com buckets customizados (inclui status para separar sucesso/falha)
	metrics.operationDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.card.operation.duration.seconds",
		"Duration of card operations in seconds by status",
		"s",
		[]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	)

	// UpDownCounter: Cartões ativos (incrementa +1 em create, decrementa -1 em delete)
	metrics.activeCardsTotal = o11y.Metrics().UpDownCounter(
		"financial.card.active.total",
		"Active cards counter (incremented on create, decremented on delete)",
		"card",
	)

	return metrics
}

// RecordOperation registra uma operação bem-sucedida
func (m *CardMetrics) RecordOperation(ctx context.Context, operation string, duration time.Duration) {
	m.operationsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("status", "success"),
	)
	m.operationDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("status", "success"),
	)
}

// RecordOperationFailure registra uma operação com falha incluindo tipo de erro
func (m *CardMetrics) RecordOperationFailure(ctx context.Context, operation string, duration time.Duration, errorType string) {
	m.operationsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("status", "failure"),
		observability.String("error_type", errorType),
	)
	m.operationDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("status", "failure"),
	)
}

// IncActiveCards incrementa o contador de cartões ativos
func (m *CardMetrics) IncActiveCards(ctx context.Context) {
	m.activeCardsTotal.Add(ctx, 1)
}

// DecActiveCards decrementa o contador de cartões ativos
func (m *CardMetrics) DecActiveCards(ctx context.Context) {
	m.activeCardsTotal.Add(ctx, -1)
}

// Constantes para tipos de operação
const (
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationFind   = "find"
	OperationFindBy = "find_by"
)

// Constantes para tipos de erro
const (
	ErrorTypeValidation = "validation"
	ErrorTypeNotFound   = "not_found"
	ErrorTypeRepository = "repository"
	ErrorTypeUnknown    = "unknown"
	ErrorTypeParsing    = "parsing"
)
