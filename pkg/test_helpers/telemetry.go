package test_helpers

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

// MockTelemetry é um mock simplificado de telemetry para testes.
type MockTelemetry struct{}

func NewMockTelemetry() *MockTelemetry {
	return &MockTelemetry{}
}

func (m *MockTelemetry) Tracer() o11y.Tracer                { return &mockTracer{} }
func (m *MockTelemetry) Logger() o11y.Logger                { return &mockLogger{} }
func (m *MockTelemetry) Metrics() o11y.Metrics              { return &mockMetrics{} }
func (m *MockTelemetry) IsClosed() bool                     { return false }
func (m *MockTelemetry) Close(ctx context.Context) error    { return nil }
func (m *MockTelemetry) Shutdown(ctx context.Context) error { return nil }

// mockTracer implementa o11y.Tracer.
type mockTracer struct{}

func (m *mockTracer) Start(ctx context.Context, name string, attrs ...o11y.Attribute) (context.Context, o11y.Span) {
	return ctx, &mockSpan{}
}
func (m *mockTracer) WithAttributes(ctx context.Context, attrs ...o11y.Attribute) {}

// mockSpan implementa o11y.Span.
type mockSpan struct{}

func (m *mockSpan) End()                                          {}
func (m *mockSpan) AddEvent(name string, attrs ...o11y.Attribute) {}
func (m *mockSpan) SetAttributes(attrs ...o11y.Attribute)         {}
func (m *mockSpan) SetStatus(status o11y.SpanStatus, desc string) {}
func (m *mockSpan) RecordError(err error)                         {}

// mockLogger implementa o11y.Logger
type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...o11y.Field)             {}
func (m *mockLogger) Error(ctx context.Context, err error, msg string, fields ...o11y.Field) {}
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...o11y.Field)            {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...o11y.Field)             {}
func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...o11y.Field)            {}

// mockMetrics implementa o11y.Metrics
type mockMetrics struct{}

func (m *mockMetrics) AddCounter(ctx context.Context, name string, value int64, attrs ...any)     {}
func (m *mockMetrics) AddHistogram(ctx context.Context, name string, value float64, attrs ...any) {}
func (m *mockMetrics) RecordHistogram(ctx context.Context, name string, value float64, attrs ...any) {
}
