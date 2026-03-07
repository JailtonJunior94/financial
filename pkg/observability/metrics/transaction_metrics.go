package metrics

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// TransactionMetrics groups metrics for the transaction module.
type TransactionMetrics struct {
	repositoryQueriesTotal observability.Counter
	repositoryDuration     observability.Histogram
	errorsTotal            observability.Counter
}

var transactionBuckets = []float64{0.005, 0.010, 0.025, 0.050, 0.100, 0.200, 0.500, 1.0, 2.0, 5.0}

// NewTransactionMetrics initializes OpenTelemetry metrics for the transaction module.
func NewTransactionMetrics(o11y observability.Observability) *TransactionMetrics {
	m := &TransactionMetrics{}
	m.repositoryQueriesTotal = o11y.Metrics().Counter(
		"financial.transaction.repository.queries.total",
		"Total number of transaction repository queries by operation, entity, and status",
		"operation",
	)
	m.repositoryDuration = o11y.Metrics().HistogramWithBuckets(
		"financial.transaction.repository.duration.seconds",
		"Duration of transaction repository queries in seconds",
		"s",
		transactionBuckets,
	)
	m.errorsTotal = o11y.Metrics().Counter(
		"financial.transaction.errors.total",
		"Total number of transaction errors by operation, entity, layer, and error_type",
		"operation",
	)
	return m
}

// RecordRepositoryQuery records a successful transaction repository query.
func (m *TransactionMetrics) RecordRepositoryQuery(ctx context.Context, operation, entity string, duration time.Duration) {
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

// RecordRepositoryFailure records a failed transaction repository query.
func (m *TransactionMetrics) RecordRepositoryFailure(ctx context.Context, operation, entity, errorType string, duration time.Duration) {
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
