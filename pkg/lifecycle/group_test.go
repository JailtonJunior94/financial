package lifecycle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/pkg/lifecycle"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/stretchr/testify/suite"
)

// mockService mock de Service para testes.
type mockService struct {
	name           string
	startFunc      func(ctx context.Context) error
	shutdownFunc   func(ctx context.Context) error
	startCalled    bool
	shutdownCalled bool
}

func newMockService(name string) *mockService {
	return &mockService{
		name: name,
		startFunc: func(ctx context.Context) error {
			return nil
		},
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

func (m *mockService) Start(ctx context.Context) error {
	m.startCalled = true
	return m.startFunc(ctx)
}

func (m *mockService) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return m.shutdownFunc(ctx)
}

func (m *mockService) Name() string {
	return m.name
}

// GroupTestSuite test suite para Group.
type GroupTestSuite struct {
	suite.Suite
	ctx  context.Context
	o11y observability.Observability
}

func TestGroupTestSuite(t *testing.T) {
	suite.Run(t, new(GroupTestSuite))
}

func (s *GroupTestSuite) SetupTest() {
	s.ctx = context.Background()

	// Mock observability (pode ser substituído por mock real se necessário)
	s.o11y = &mockObservability{}
}

func (s *GroupTestSuite) TestNewGroup() {
	// Arrange & Act
	group := lifecycle.NewGroup(s.o11y, nil)

	// Assert
	s.NotNil(group)
}

func (s *GroupTestSuite) TestRegister() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())
	service1 := newMockService("service-1")
	service2 := newMockService("service-2")

	// Act
	group.Register(service1)
	group.Register(service2)

	// Assert
	// Verificação implícita: não deve dar panic
	s.NotNil(group)
}

func (s *GroupTestSuite) TestStart_Success() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())
	service1 := newMockService("service-1")
	service2 := newMockService("service-2")

	group.Register(service1)
	group.Register(service2)

	// Act
	err := group.Start(s.ctx)

	// Assert
	s.NoError(err)
	s.True(service1.startCalled, "service1 Start deve ser chamado")
	s.True(service2.startCalled, "service2 Start deve ser chamado")
}

func (s *GroupTestSuite) TestStart_FailsOnFirstServiceError() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())
	service1 := newMockService("service-1")
	service2 := newMockService("service-2")

	// Configurar service1 para falhar
	expectedErr := errors.New("start failed")
	service1.startFunc = func(ctx context.Context) error {
		return expectedErr
	}

	group.Register(service1)
	group.Register(service2)

	// Act
	err := group.Start(s.ctx)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "failed to start service service-1")
	s.True(service1.startCalled, "service1 Start deve ser chamado")
	s.False(service2.startCalled, "service2 Start NÃO deve ser chamado após falha")
}

func (s *GroupTestSuite) TestStart_RespectContextTimeout() {
	// Arrange
	config := &lifecycle.GroupConfig{
		StartTimeout:    100 * time.Millisecond,
		ShutdownTimeout: 1 * time.Second,
	}
	group := lifecycle.NewGroup(s.o11y, config)
	service := newMockService("slow-service")

	// Configurar service para demorar mais que o timeout
	service.startFunc = func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	group.Register(service)

	// Act
	err := group.Start(s.ctx)

	// Assert
	s.Error(err)
	s.True(service.startCalled)
}

func (s *GroupTestSuite) TestShutdown_Success() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())
	service1 := newMockService("service-1")
	service2 := newMockService("service-2")

	group.Register(service1)
	group.Register(service2)

	// Start services primeiro
	err := group.Start(s.ctx)
	s.NoError(err)

	// Act
	err = group.Shutdown(s.ctx)

	// Assert
	s.NoError(err)
	s.True(service1.shutdownCalled, "service1 Shutdown deve ser chamado")
	s.True(service2.shutdownCalled, "service2 Shutdown deve ser chamado")
}

func (s *GroupTestSuite) TestShutdown_Parallel() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())

	service1 := newMockService("service-1")
	service2 := newMockService("service-2")
	service3 := newMockService("service-3")

	group.Register(service1)
	group.Register(service2)
	group.Register(service3)

	// Act
	err := group.Shutdown(s.ctx)

	// Assert
	s.NoError(err)
	// Shutdown é paralelo, verificamos que todos foram chamados
	s.True(service1.shutdownCalled, "service1 Shutdown deve ser chamado")
	s.True(service2.shutdownCalled, "service2 Shutdown deve ser chamado")
	s.True(service3.shutdownCalled, "service3 Shutdown deve ser chamado")
}

func (s *GroupTestSuite) TestShutdown_WithErrors() {
	// Arrange
	group := lifecycle.NewGroup(s.o11y, lifecycle.DefaultGroupConfig())
	service1 := newMockService("service-1")
	service2 := newMockService("service-2")

	// Configurar service2 para falhar
	expectedErr := errors.New("shutdown failed")
	service2.shutdownFunc = func(ctx context.Context) error {
		return expectedErr
	}

	group.Register(service1)
	group.Register(service2)

	// Act
	err := group.Shutdown(s.ctx)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "shutdown errors")
	s.True(service1.shutdownCalled, "service1 Shutdown deve ser chamado")
	s.True(service2.shutdownCalled, "service2 Shutdown deve ser chamado")
}

func (s *GroupTestSuite) TestShutdown_Timeout() {
	// Arrange
	config := &lifecycle.GroupConfig{
		StartTimeout:    1 * time.Second,
		ShutdownTimeout: 100 * time.Millisecond,
	}
	group := lifecycle.NewGroup(s.o11y, config)
	service := newMockService("slow-shutdown-service")

	// Configurar service para demorar mais que o timeout
	service.shutdownFunc = func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	group.Register(service)

	// Act
	err := group.Shutdown(s.ctx)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "shutdown timeout exceeded")
}

// mockObservability mock simples de observability.
type mockObservability struct{}

func (m *mockObservability) Logger() observability.Logger {
	return &mockLogger{}
}

func (m *mockObservability) Tracer() observability.Tracer {
	return &mockTracer{}
}

func (m *mockObservability) Metrics() observability.Metrics {
	return nil
}

func (m *mockObservability) Shutdown(ctx context.Context) error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...observability.Field)  {}
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...observability.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {}
func (m *mockLogger) With(fields ...observability.Field) observability.Logger              { return m }

type mockTracer struct{}

func (m *mockTracer) Start(ctx context.Context, spanName string, opts ...observability.SpanOption) (context.Context, observability.Span) {
	return ctx, &mockSpan{}
}

func (m *mockTracer) ContextWithSpan(ctx context.Context, span observability.Span) context.Context {
	return ctx
}

func (m *mockTracer) SpanFromContext(ctx context.Context) observability.Span {
	return &mockSpan{}
}

type mockSpan struct{}

func (m *mockSpan) End()                                                 {}
func (m *mockSpan) AddEvent(name string, fields ...observability.Field)  {}
func (m *mockSpan) SetAttributes(attrs ...observability.Field)           {}
func (m *mockSpan) RecordError(err error, fields ...observability.Field) {}
func (m *mockSpan) SetStatus(code observability.StatusCode, msg string)  {}
func (m *mockSpan) Context() observability.SpanContext                   { return nil }
