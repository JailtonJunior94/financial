package metrics

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// FinancialMetrics agrupa todas as métricas transversais da aplicação financeira.
type FinancialMetrics struct {
	o11y observability.Observability

	// Usecase metrics
	usecaseOperationsTotal observability.Counter
	usecaseDuration        observability.Histogram

	// Repository metrics
	repositoryQueriesTotal observability.Counter
	repositoryDuration     observability.Histogram

	// Handler metrics
	handlerRequestsTotal observability.Counter
	handlerDuration      observability.Histogram

	// External call metrics
	externalCallsTotal observability.Counter
	externalDuration   observability.Histogram

	// Consumer metrics
	consumerEventsTotal observability.Counter

	// Errors total
	errorsTotal observability.Counter
}

var financialBuckets = []float64{0.005, 0.010, 0.025, 0.050, 0.100, 0.200, 0.500, 1.0, 2.0, 5.0}

// NewFinancialMetrics inicializa métricas OpenTelemetry para a aplicação financeira.
func NewFinancialMetrics(o11y observability.Observability) *FinancialMetrics {
	m := &FinancialMetrics{o11y: o11y}

	m.usecaseOperationsTotal = o11y.Metrics().Counter(
		"financial.usecase.operations.total",
		"Total number of usecase operations by operation, entity, and status",
		"operation",
	)
	m.usecaseDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.usecase.duration.seconds",
		"Duration of usecase operations in seconds",
		"s",
		financialBuckets,
	)

	m.repositoryQueriesTotal = o11y.Metrics().Counter(
		"financial.repository.queries.total",
		"Total number of repository queries by operation, entity, and status",
		"operation",
	)
	m.repositoryDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.repository.duration.seconds",
		"Duration of repository queries in seconds",
		"s",
		financialBuckets,
	)

	m.handlerRequestsTotal = o11y.Metrics().Counter(
		"financial.handler.requests.total",
		"Total number of HTTP handler requests by operation, entity, and status",
		"operation",
	)
	m.handlerDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.handler.duration.seconds",
		"Duration of HTTP handler requests in seconds",
		"s",
		financialBuckets,
	)

	m.externalCallsTotal = o11y.Metrics().Counter(
		"financial.external.calls.total",
		"Total number of external calls by operation, entity, and status",
		"operation",
	)
	m.externalDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.external.duration.seconds",
		"Duration of external calls in seconds",
		"s",
		financialBuckets,
	)

	m.consumerEventsTotal = o11y.Metrics().Counter(
		"financial.consumer.events.total",
		"Total number of consumer events processed by operation, entity, and status",
		"operation",
	)

	m.errorsTotal = o11y.Metrics().Counter(
		"financial.errors.total",
		"Total number of errors by operation, entity, layer, and error_type",
		"operation",
	)

	return m
}

// RecordUsecaseOperation registra uma operação de usecase bem-sucedida.
func (m *FinancialMetrics) RecordUsecaseOperation(ctx context.Context, operation, entity string, duration time.Duration) {
	m.usecaseOperationsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
	m.usecaseDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
}

// RecordUsecaseFailure registra uma operação de usecase com falha.
func (m *FinancialMetrics) RecordUsecaseFailure(ctx context.Context, operation, entity, errorType string, duration time.Duration) {
	m.usecaseOperationsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.usecaseDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.errorsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("layer", "usecase"),
		observability.String("error_type", errorType),
	)
}

// RecordRepositoryQuery registra uma query de repositório bem-sucedida.
func (m *FinancialMetrics) RecordRepositoryQuery(ctx context.Context, operation, entity string, duration time.Duration) {
	m.repositoryQueriesTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
	m.repositoryDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
}

// RecordRepositoryFailure registra uma query de repositório com falha.
func (m *FinancialMetrics) RecordRepositoryFailure(ctx context.Context, operation, entity, errorType string, duration time.Duration) {
	m.repositoryQueriesTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.repositoryDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.errorsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("layer", "repository"),
		observability.String("error_type", errorType),
	)
}

// RecordHandlerRequest registra uma requisição de handler bem-sucedida.
func (m *FinancialMetrics) RecordHandlerRequest(ctx context.Context, operation, entity string, duration time.Duration) {
	m.handlerRequestsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
	m.handlerDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
}

// RecordHandlerFailure registra uma requisição de handler com falha.
func (m *FinancialMetrics) RecordHandlerFailure(ctx context.Context, operation, entity, errorType string, duration time.Duration) {
	m.handlerRequestsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.handlerDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.errorsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("layer", "handler"),
		observability.String("error_type", errorType),
	)
}

// RecordExternalCall registra uma chamada externa bem-sucedida.
func (m *FinancialMetrics) RecordExternalCall(ctx context.Context, operation, entity string, duration time.Duration) {
	m.externalCallsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
	m.externalDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
}

// RecordExternalFailure registra uma chamada externa com falha.
func (m *FinancialMetrics) RecordExternalFailure(ctx context.Context, operation, entity, errorType string, duration time.Duration) {
	m.externalCallsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.externalDuration.Record(ctx, duration.Seconds(),
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.errorsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("layer", "external"),
		observability.String("error_type", errorType),
	)
}

// RecordConsumerEvent registra um evento de consumer processado com sucesso.
func (m *FinancialMetrics) RecordConsumerEvent(ctx context.Context, operation, entity string) {
	m.consumerEventsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "success"),
	)
}

// RecordConsumerFailure registra um evento de consumer com falha.
func (m *FinancialMetrics) RecordConsumerFailure(ctx context.Context, operation, entity, errorType string) {
	m.consumerEventsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("status", "failure"),
	)
	m.errorsTotal.Increment(ctx,
		observability.String("operation", operation),
		observability.String("entity", entity),
		observability.String("layer", "consumer"),
		observability.String("error_type", errorType),
	)
}

// Constantes de operação adicionais (complementam as definidas em card_metrics.go)
const (
	OperationGet        = "get"
	OperationFindByCode = "find_by_code"
	OperationToken      = "token"
	OperationList       = "list"
	OperationSync       = "sync"
	OperationRegister   = "register"
)

// Constantes de entidade
const (
	EntityBudgetItem = "budget_item"
)
