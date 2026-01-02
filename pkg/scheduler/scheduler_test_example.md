# Exemplo de Testes para o Scheduler

Este arquivo demonstra como criar testes unitários para jobs e o scheduler.

## Estrutura de Teste Sugerida

```
pkg/scheduler/
├── scheduler.go
└── scheduler_test.go         # Testes do scheduler

internal/worker/jobs/
├── database_cleanup_job.go
├── database_cleanup_job_test.go
├── report_generator_job.go
└── report_generator_job_test.go
```

## Exemplo 1: Testando um Job (Unit Test)

```go
// internal/worker/jobs/database_cleanup_job_test.go
package jobs_test

import (
    "context"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jailtonjunior94/financial/internal/worker/jobs"
    "github.com/stretchr/testify/suite"
)

type DatabaseCleanupJobSuite struct {
    suite.Suite
    ctx  context.Context
    mock sqlmock.Sqlmock
}

func TestDatabaseCleanupJobSuite(t *testing.T) {
    suite.Run(t, new(DatabaseCleanupJobSuite))
}

func (s *DatabaseCleanupJobSuite) SetupTest() {
    s.ctx = context.Background()
    
    db, mock, err := sqlmock.New()
    s.Require().NoError(err)
    s.mock = mock
}

func (s *DatabaseCleanupJobSuite) TestRun_Success() {
    // Arrange
    cutoffDate := time.Now().Add(-90 * 24 * time.Hour)
    
    s.mock.ExpectExec("DELETE FROM categories").
        WithArgs(sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(0, 42)) // 42 rows deleted
    
    // Mock observability (ou use uma implementação fake)
    // o11y := &fakeO11y{}
    
    // job := jobs.NewDatabaseCleanupJob(db, o11y)
    
    // Act
    // err := job.Run(s.ctx)
    
    // Assert
    // s.NoError(err)
    // s.NoError(s.mock.ExpectationsWereMet())
}

func (s *DatabaseCleanupJobSuite) TestRun_DatabaseError() {
    // Arrange
    s.mock.ExpectExec("DELETE FROM categories").
        WillReturnError(fmt.Errorf("database connection lost"))
    
    // Act
    // err := job.Run(s.ctx)
    
    // Assert
    // s.Error(err)
    // s.Contains(err.Error(), "database connection lost")
}

func (s *DatabaseCleanupJobSuite) TestName() {
    // job := jobs.NewDatabaseCleanupJob(db, o11y)
    // s.Equal("database_cleanup", job.Name())
}

func (s *DatabaseCleanupJobSuite) TestSchedule() {
    // job := jobs.NewDatabaseCleanupJob(db, o11y)
    // s.Equal("0 2 * * *", job.Schedule())
}
```

## Exemplo 2: Testando o Scheduler

```go
// pkg/scheduler/scheduler_test.go
package scheduler_test

import (
    "context"
    "errors"
    "sync/atomic"
    "testing"
    "time"

    "github.com/jailtonjunior94/financial/pkg/jobs"
    "github.com/jailtonjunior94/financial/pkg/scheduler"
    "github.com/stretchr/testify/suite"
)

// Mock Job for testing
type mockJob struct {
    name        string
    schedule    string
    runFunc     func(ctx context.Context) error
    runCount    atomic.Int32
}

func (m *mockJob) Name() string     { return m.name }
func (m *mockJob) Schedule() string { return m.schedule }
func (m *mockJob) Run(ctx context.Context) error {
    m.runCount.Add(1)
    if m.runFunc != nil {
        return m.runFunc(ctx)
    }
    return nil
}

type SchedulerSuite struct {
    suite.Suite
    ctx    context.Context
    cancel context.CancelFunc
}

func TestSchedulerSuite(t *testing.T) {
    suite.Run(t, new(SchedulerSuite))
}

func (s *SchedulerSuite) SetupTest() {
    s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *SchedulerSuite) TearDownTest() {
    s.cancel()
}

func (s *SchedulerSuite) TestRegister_ValidJob() {
    // Arrange
    // o11y := &fakeO11y{}
    config := &jobs.Config{
        DefaultTimeout:    5 * time.Second,
        EnableRecovery:    true,
        MaxConcurrentJobs: 5,
    }
    // sched := scheduler.New(s.ctx, o11y, config)
    
    job := &mockJob{
        name:     "test_job",
        schedule: "@every 1s",
    }
    
    // Act
    // err := sched.Register(job)
    
    // Assert
    // s.NoError(err)
}

func (s *SchedulerSuite) TestRegister_InvalidSchedule() {
    // Arrange
    // sched := scheduler.New(s.ctx, o11y, config)
    
    job := &mockJob{
        name:     "invalid_job",
        schedule: "INVALID CRON",
    }
    
    // Act
    // err := sched.Register(job)
    
    // Assert
    // s.Error(err)
    // s.Contains(err.Error(), "failed to register job")
}

func (s *SchedulerSuite) TestJobExecution() {
    // Arrange
    executed := make(chan struct{})
    
    job := &mockJob{
        name:     "quick_job",
        schedule: "@every 100ms",
        runFunc: func(ctx context.Context) error {
            select {
            case executed <- struct{}{}:
            default:
            }
            return nil
        },
    }
    
    // sched := scheduler.New(s.ctx, o11y, config)
    // sched.Register(job)
    // sched.Start()
    
    // Act & Assert
    // Wait for at least one execution
    // select {
    // case <-executed:
    //     s.GreaterOrEqual(job.runCount.Load(), int32(1))
    // case <-time.After(1 * time.Second):
    //     s.Fail("job did not execute in time")
    // }
}

func (s *SchedulerSuite) TestRecoveryOnPanic() {
    // Arrange
    panicJob := &mockJob{
        name:     "panic_job",
        schedule: "@every 100ms",
        runFunc: func(ctx context.Context) error {
            panic("intentional panic for testing")
        },
    }
    
    normalJob := &mockJob{
        name:     "normal_job",
        schedule: "@every 100ms",
    }
    
    // sched := scheduler.New(s.ctx, o11y, config)
    // sched.Register(panicJob)
    // sched.Register(normalJob)
    // sched.Start()
    
    // Act
    // time.Sleep(500 * time.Millisecond)
    
    // Assert
    // Normal job should still execute despite panic in other job
    // s.GreaterOrEqual(normalJob.runCount.Load(), int32(1))
}

func (s *SchedulerSuite) TestConcurrencyLimit() {
    // Arrange
    config := &jobs.Config{
        DefaultTimeout:    10 * time.Second,
        MaxConcurrentJobs: 2, // Only 2 concurrent executions
    }
    
    executing := make(chan struct{}, 10)
    block := make(chan struct{})
    
    job := &mockJob{
        name:     "slow_job",
        schedule: "@every 10ms", // Very frequent
        runFunc: func(ctx context.Context) error {
            executing <- struct{}{}
            <-block // Block until released
            return nil
        },
    }
    
    // sched := scheduler.New(s.ctx, o11y, config)
    // sched.Register(job)
    // sched.Start()
    
    // Act
    // time.Sleep(100 * time.Millisecond) // Allow multiple triggers
    
    // Assert
    // Should have max 2 executing due to limit
    // s.LessOrEqual(len(executing), 2)
    
    // Cleanup
    // close(block)
}

func (s *SchedulerSuite) TestGracefulShutdown() {
    // Arrange
    jobStarted := make(chan struct{})
    jobCanFinish := make(chan struct{})
    jobFinished := make(chan struct{})
    
    job := &mockJob{
        name:     "long_job",
        schedule: "@every 1s",
        runFunc: func(ctx context.Context) error {
            close(jobStarted)
            <-jobCanFinish
            close(jobFinished)
            return nil
        },
    }
    
    // sched := scheduler.New(s.ctx, o11y, config)
    // sched.Register(job)
    // sched.Start()
    
    // Wait for job to start
    // <-jobStarted
    
    // Act - Initiate shutdown
    // shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    // defer cancel()
    
    // go func() {
    //     time.Sleep(100 * time.Millisecond)
    //     close(jobCanFinish) // Allow job to finish
    // }()
    
    // err := sched.Shutdown(shutdownCtx)
    
    // Assert
    // s.NoError(err)
    
    // Wait for job to complete
    // select {
    // case <-jobFinished:
    //     // Job completed successfully
    // case <-time.After(1 * time.Second):
    //     s.Fail("job did not finish")
    // }
}

func (s *SchedulerSuite) TestShutdownTimeout() {
    // Arrange
    config := &jobs.Config{
        DefaultTimeout:    30 * time.Second, // Long job timeout
    }
    
    blockForever := make(chan struct{})
    
    job := &mockJob{
        name:     "stuck_job",
        schedule: "@every 1s",
        runFunc: func(ctx context.Context) error {
            <-blockForever // Never completes
            return nil
        },
    }
    
    // sched := scheduler.New(s.ctx, o11y, config)
    // sched.Register(job)
    // sched.Start()
    
    // time.Sleep(100 * time.Millisecond) // Allow job to start
    
    // Act
    // shutdownCtx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    // defer cancel()
    
    // err := sched.Shutdown(shutdownCtx)
    
    // Assert
    // Should timeout because job never finishes
    // s.Error(err)
    // s.Contains(err.Error(), "timeout")
}
```

## Exemplo 3: Fake Observability para Testes

```go
// internal/worker/jobs/testing/fake_observability.go
package testing

import (
    "context"
    "github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type FakeLogger struct {
    InfoCalls  []LogCall
    ErrorCalls []LogCall
    WarnCalls  []LogCall
}

type LogCall struct {
    Message string
    Fields  []observability.Field
}

func (f *FakeLogger) Info(ctx context.Context, msg string, fields ...observability.Field) {
    f.InfoCalls = append(f.InfoCalls, LogCall{Message: msg, Fields: fields})
}

func (f *FakeLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {
    f.ErrorCalls = append(f.ErrorCalls, LogCall{Message: msg, Fields: fields})
}

func (f *FakeLogger) Warn(ctx context.Context, msg string, fields ...observability.Field) {
    f.WarnCalls = append(f.WarnCalls, LogCall{Message: msg, Fields: fields})
}

func (f *FakeLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {}

type FakeObservability struct {
    logger *FakeLogger
}

func NewFakeObservability() (*FakeObservability, *FakeLogger) {
    logger := &FakeLogger{}
    return &FakeObservability{logger: logger}, logger
}

func (f *FakeObservability) Logger() observability.Logger {
    return f.logger
}

func (f *FakeObservability) Tracer() observability.Tracer {
    return nil // Não necessário para testes básicos
}

func (f *FakeObservability) Metrics() observability.Metrics {
    return nil
}

func (f *FakeObservability) Shutdown(ctx context.Context) error {
    return nil
}
```

## Exemplo 4: Teste de Integração com Testcontainers

```go
// internal/worker/jobs/database_cleanup_job_integration_test.go
// +build integration

package jobs_test

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "github.com/jailtonjunior94/financial/internal/worker/jobs"
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

type DatabaseCleanupIntegrationSuite struct {
    suite.Suite
    ctx       context.Context
    db        *sql.DB
    container testcontainers.Container
}

func TestDatabaseCleanupIntegrationSuite(t *testing.T) {
    suite.Run(t, new(DatabaseCleanupIntegrationSuite))
}

func (s *DatabaseCleanupIntegrationSuite) SetupSuite() {
    s.ctx = context.Background()
    
    // Start PostgreSQL container
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }
    
    container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    s.Require().NoError(err)
    s.container = container
    
    // Connect to database
    host, _ := container.Host(s.ctx)
    port, _ := container.MappedPort(s.ctx, "5432")
    
    dsn := fmt.Sprintf("host=%s port=%s user=postgres password=test dbname=testdb sslmode=disable",
        host, port.Port())
    
    db, err := sql.Open("postgres", dsn)
    s.Require().NoError(err)
    s.db = db
    
    // Create test table
    _, err = s.db.Exec(`
        CREATE TABLE categories (
            id UUID PRIMARY KEY,
            name VARCHAR(255),
            deleted_at TIMESTAMP
        )
    `)
    s.Require().NoError(err)
}

func (s *DatabaseCleanupIntegrationSuite) TearDownSuite() {
    s.db.Close()
    s.container.Terminate(s.ctx)
}

func (s *DatabaseCleanupIntegrationSuite) TestCleanupOldRecords() {
    // Arrange - Insert test data
    oldDate := time.Now().Add(-100 * 24 * time.Hour) // 100 days ago
    recentDate := time.Now().Add(-10 * 24 * time.Hour) // 10 days ago
    
    s.db.Exec("INSERT INTO categories VALUES ($1, 'Old', $2)",
        "old-uuid", oldDate)
    s.db.Exec("INSERT INTO categories VALUES ($1, 'Recent', $2)",
        "recent-uuid", recentDate)
    
    // Act
    // o11y, _ := NewFakeObservability()
    // job := jobs.NewDatabaseCleanupJob(s.db, o11y)
    // err := job.Run(s.ctx)
    
    // Assert
    // s.NoError(err)
    
    // Verify old record deleted, recent kept
    // var count int
    // s.db.QueryRow("SELECT COUNT(*) FROM categories WHERE id = 'old-uuid'").Scan(&count)
    // s.Equal(0, count, "Old record should be deleted")
    
    // s.db.QueryRow("SELECT COUNT(*) FROM categories WHERE id = 'recent-uuid'").Scan(&count)
    // s.Equal(1, count, "Recent record should be kept")
}
```

## Executando Testes

```bash
# Testes unitários
go test -v ./pkg/scheduler/...
go test -v ./internal/worker/jobs/...

# Testes de integração
go test -v -tags=integration ./internal/worker/jobs/...

# Com cobertura
go test -v -coverprofile=coverage.out ./pkg/scheduler/...
go tool cover -html=coverage.out
```

## Ferramentas Recomendadas

- `testify/suite`: Organização de testes
- `testify/assert`: Asserções
- `testify/mock`: Mocks (se necessário)
- `go-sqlmock`: Mock de database
- `testcontainers-go`: Testes de integração
- `ginkgo/gomega`: Framework BDD (alternativa)

## Boas Práticas

1. **Testes isolados**: Cada teste independente
2. **Nomes descritivos**: `TestJobName_Scenario_ExpectedBehavior`
3. **AAA Pattern**: Arrange, Act, Assert
4. **Table-driven tests**: Para múltiplos cenários
5. **Cleanup**: Sempre limpar recursos (defer)
6. **Context**: Usar context com timeout nos testes
7. **Parallel**: Marcar testes que podem rodar em paralelo (`t.Parallel()`)
