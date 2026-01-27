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
	errorsTotal       observability.Counter
	operationDuration observability.Histogram
	activeCardsTotal  observability.UpDownCounter
}

// NewCardMetrics inicializa métricas OpenTelemetry para o módulo de cartões
func NewCardMetrics(o11y observability.Observability) *CardMetrics {
	metrics := &CardMetrics{
		o11y: o11y,
	}

	// Counter: Operações executadas
	metrics.operationsTotal = o11y.Metrics().Counter(
		"financial.card.operations.total",
		"Total number of card operations executed by type (create, update, delete, find, find_by)",
		"operation",
	)

	// Counter: Erros detalhados
	metrics.errorsTotal = o11y.Metrics().Counter(
		"financial.card.errors.total",
		"Total number of errors by operation and error type",
		"error",
	)

	// Histogram: Latência com buckets customizados
	metrics.operationDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.card.operation.duration.seconds",
		"Duration of card operations in seconds",
		"s",
		[]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	)

	// UpDownCounter: Cartões ativos (substitui Gauge)
	metrics.activeCardsTotal = o11y.Metrics().UpDownCounter(
		"financial.card.active.total",
		"Current number of active cards in the system",
		"card",
	)

	return metrics
}

// RecordOperation registra uma operação bem-sucedida
func (m *CardMetrics) RecordOperation(ctx context.Context, operation string, duration time.Duration) {
	m.operationsTotal.Increment(ctx,
		observability.Field{Key: "operation", Value: operation},
		observability.Field{Key: "status", Value: "success"},
	)
	m.operationDuration.Record(ctx, duration.Seconds(),
		observability.Field{Key: "operation", Value: operation},
	)
}

// RecordOperationFailure registra uma operação com falha
func (m *CardMetrics) RecordOperationFailure(ctx context.Context, operation string, duration time.Duration) {
	m.operationsTotal.Increment(ctx,
		observability.Field{Key: "operation", Value: operation},
		observability.Field{Key: "status", Value: "failure"},
	)
	m.operationDuration.Record(ctx, duration.Seconds(),
		observability.Field{Key: "operation", Value: operation},
	)
}

// RecordError registra um erro específico
func (m *CardMetrics) RecordError(ctx context.Context, operation, errorType string) {
	m.errorsTotal.Increment(ctx,
		observability.Field{Key: "operation", Value: operation},
		observability.Field{Key: "error_type", Value: errorType},
	)
}

// SetActiveCards atualiza o total de cartões ativos
func (m *CardMetrics) SetActiveCards(ctx context.Context, count float64) {
	// UpDownCounter não tem Set(), mas Add() funciona para ajustes
	m.activeCardsTotal.Add(ctx, int64(count))
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
