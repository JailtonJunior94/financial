# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Environment Setup
```bash
make dotenv  # Generate .env file from .env.example in cmd/
```

### Build
```bash
make build   # Compiles to ./bin/financial (CGO_ENABLED=0, optimized with -ldflags="-w -s")
```

### Testing
```bash
make test    # Run all tests with race detection and coverage
make cover   # Generate and view HTML coverage report
```

### Linting
```bash
make lint    # Run golangci-lint with .golangci.yml configuration
```

### Mocks
```bash
make mocks   # Generate mocks using mockery (configured in .mockery.yml)
```

### Database Migrations
```bash
make migrate NAME=migration_name  # Create new migration files in database/migrations
```

### Running the Application
```bash
# Run API server
./bin/financial api

# Run database migrations (CockroachDB)
./bin/financial migrate

# Consumers (not yet implemented)
./bin/financial consumers
```

### Docker
```bash
make start_minimal  # Start CockroachDB, migration, RabbitMQ, and OTEL
make start_docker   # Start all services
make stop_docker    # Stop all services
```

## Architecture

### Clean Architecture with Domain-Driven Design

The codebase follows Clean Architecture principles organized into distinct layers:

**Domain Layer** (`internal/{module}/domain/`)
- `entities/`: Core business entities with behavior (e.g., Budget, Category, User)
- `vos/`: Value Objects (e.g., CategoryName, Email, Money, Percentage)
- `factories/`: Entity creation with validation logic
- `interfaces/`: Repository interfaces (dependency inversion)

**Application Layer** (`internal/{module}/application/`)
- `usecase/`: Application-specific business rules and orchestration
- `dtos/`: Data Transfer Objects for request/response

**Infrastructure Layer** (`internal/{module}/infrastructure/`)
- `repositories/`: Database implementations of domain interfaces
- `http/`: HTTP handlers, routes, and transport concerns
- `mocks/`: Generated mocks for testing (via mockery)

**Module Registration Pattern**
Each domain module (user, category, budget) has a `module.go` file that:
- Wires up dependencies (repositories, use cases, handlers)
- Returns HTTP routes for registration
- Encapsulates module initialization

### Dependency Injection Container

The `pkg/bundle/container.go` provides centralized dependency management:
- Database connection (`*sql.DB`)
- Configuration (`configs.Config`)
- JWT and hashing adapters
- OpenTelemetry telemetry (traces, metrics, logs via devkit-go)
- Shared middlewares (auth, panic recovery)

### Unit of Work Pattern

`pkg/database/uow/` implements Unit of Work for transactional operations:
- `Executor()`: Returns current DB executor (transaction or connection)
- `Do(ctx, fn)`: Wraps operations in a transaction with automatic rollback
- Used primarily in Budget module for multi-entity operations

### Database Strategy

- **Primary Database**: CockroachDB (PostgreSQL-compatible)
- **Support**: MySQL and Postgres drivers available
- **Migrations**: Uses golang-migrate with Unix timestamp naming
- **Testing**: Testcontainers for integration tests (CockroachDB and Postgres modules)

### HTTP Server Architecture

Built on `github.com/JailtonJunior94/devkit-go/pkg/httpserver`:
- Custom error handling via `pkg/custom_errors` and `pkg/httperrors`
- Middleware chain: RequestID, Auth, Panic Recovery, Metrics, Tracing
- Health check endpoint at `/health` (pings database)
- Routes registered per module with optional middleware

### Custom Error Handling

`pkg/custom_errors/custom_errors.go`:
- `CustomError` wraps errors with additional context (message, details)
- `pkg/httperrors/http_errors.go` maps domain errors to HTTP status codes
- Server extracts original error from CustomError for proper HTTP mapping
- Supports error details in JSON responses

### Authentication

JWT-based authentication (`pkg/auth/jwt.go`):
- Token generation with configurable duration
- Middleware validates tokens and extracts user claims
- Configuration via `AUTH_SECRET_KEY` and `AUTH_TOKEN_DURATION`

### Observability

Full OpenTelemetry integration via devkit-go:
- Distributed tracing (OTLP gRPC)
- Metrics collection
- Structured logging
- Configure via `OTEL_*` environment variables
- Middleware automatically instruments HTTP handlers

### Testing Strategy

- Domain entities have unit tests (`*_test.go`)
- Repository tests use testcontainers for real database scenarios
- Mocks generated via mockery for use case testing
- Coverage reports generated with `make cover`

## Testing Patterns

### AAA Pattern (Arrange-Act-Assert)

All tests should follow the AAA (Arrange-Act-Assert) pattern for clarity and maintainability:

- **Arrange**: Set up test data, mocks, and dependencies
- **Act**: Execute the function/method being tested
- **Assert**: Verify the expected outcomes

### Unit Testing with Mocks

For unit tests, always use the `.mockery.yml` configuration and generate mocks with:

```bash
make mocks
```

This ensures consistent mock generation across the codebase.

### Test Structure with testify/suite

Use `testify/suite` for organizing related tests with shared setup:

```go
package usecase_test

import (
    "context"
    "errors"
    "testing"

    "your-project/internal/domain/interfaces/mocks"
    "your-project/internal/application/usecase"
    "your-project/internal/application/dtos"

    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type CreateCategoryUseCaseSuite struct {
    suite.Suite

    ctx                context.Context
    categoryRepository *mocks.CategoryRepository
}

func TestCreateCategoryUseCaseSuite(t *testing.T) {
    suite.Run(t, new(CreateCategoryUseCaseSuite))
}

func (s *CreateCategoryUseCaseSuite) SetupTest() {
    s.ctx = context.Background()
    s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *CreateCategoryUseCaseSuite) TestExecute() {
    type (
        args struct {
            input *dtos.CreateCategoryInput
        }

        dependencies struct {
            categoryRepository *mocks.CategoryRepository
        }
    )

    scenarios := []struct {
        name         string
        args         args
        dependencies dependencies
        expect       func(output *dtos.CategoryOutput, err error)
    }{
        {
            name: "deve criar uma nova categoria com sucesso",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(nil).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar resultados
                s.NoError(err)
                s.NotNil(output)
                s.Equal("Transport", output.Name)
            },
        },
        {
            name: "deve retornar erro ao falhar ao criar categoria",
            args: args{
                input: &dtos.CreateCategoryInput{
                    Name: "Transport",
                },
            },
            dependencies: dependencies{
                categoryRepository: func() *mocks.CategoryRepository {
                    // Arrange: configurar mock para falhar
                    s.categoryRepository.
                        EXPECT().
                        Create(s.ctx, mock.Anything).
                        Return(errors.New("database error")).
                        Once()
                    return s.categoryRepository
                }(),
            },
            expect: func(output *dtos.CategoryOutput, err error) {
                // Assert: verificar erro
                s.Error(err)
                s.Nil(output)
                s.Contains(err.Error(), "database error")
            },
        },
    }

    for _, scenario := range scenarios {
        s.T().Run(scenario.name, func(t *testing.T) {
            // Act: executar o use case
            uc := usecase.NewCreateCategoryUseCase(scenario.dependencies.categoryRepository)
            output, err := uc.Execute(s.ctx, scenario.args.input)

            // Assert: chamar função de verificação
            scenario.expect(output, err)
        })
    }
}
```

### Key Testing Principles

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Table-Driven Tests**: Use scenario-based approach for multiple test cases
3. **Mock Configuration**: Configure mocks inside dependency functions for each scenario
4. **Clear Naming**: Use descriptive test names in Portuguese explaining the expected behavior
5. **Comprehensive Coverage**: Test both success and error scenarios
6. **Mock Verification**: Use `.Once()` to ensure mocks are called exactly once
7. **Setup and Teardown**: Use `SetupTest()` for test initialization

### Running Specific Tests

```bash
# Run all tests in a package
go test -v -count=1 ./internal/category/application/usecase/

# Run a specific test suite
go test -v -count=1 -run TestCreateCategoryUseCaseSuite ./internal/category/application/usecase/

# Run a specific test case
go test -v -count=1 -run TestCreateCategoryUseCaseSuite/deve_criar_uma_nova_categoria ./internal/category/application/usecase/
```

## Project Structure

```
.
├── cmd/
│   ├── main.go           # Cobra CLI entry (api, migrate, consumers commands)
│   ├── server/           # HTTP server setup and module wiring
│   └── .env.example      # Environment template
├── internal/
│   ├── user/             # User domain (auth, creation)
│   ├── category/         # Category domain (hierarchical with children)
│   └── budget/           # Budget domain (with items, percentage validation)
├── pkg/
│   ├── bundle/           # Dependency injection container
│   ├── database/         # DB abstraction, migrations, UoW, test helpers
│   ├── auth/             # JWT implementation
│   ├── api/              # HTTP utilities, middlewares, error handlers
│   ├── custom_errors/    # Error wrapping and mapping
│   └── linq/             # Utility functions for slices
├── configs/              # Viper-based configuration loading
├── database/migrations/  # SQL migrations (Unix timestamp format)
└── deployment/           # Docker Compose and container configs
```

## Configuration

Configuration loaded from `.env` via Viper (see `cmd/.env.example`):
- HTTP port
- Database connection (supports Postgres/CockroachDB/MySQL)
- OpenTelemetry endpoints
- JWT secret and token duration

## Important Patterns

### Value Objects
Extensively used for domain validation (Money, Percentage, Email, CategoryName).
Created via factory functions that enforce business rules.

### Repository Pattern
All persistence abstracted behind interfaces in domain layer.
Repositories accept `database.DBExecutor` interface (supports both `*sql.DB` and `*sql.Tx`).

### Budget Domain Rules
- Budget must have items with percentages totaling exactly 100%
- `AddItem()` and `AddItems()` return `bool` indicating if percentage rule is satisfied
- AmountUsed and PercentageUsed calculated automatically when items added

### Category Hierarchy
Categories support parent-child relationships via `ParentID *UUID`.
Repository loads children recursively and attaches via `AddChildrens()`.

### Soft Deletes
Entities use `DeletedAt` timestamps (via `sharedVos.NullableTime`).
Call entity's `Delete()` method to set timestamp.

## Running Tests for a Single Package

```bash
go test -v -count=1 ./internal/category/domain/entities/
go test -v -count=1 ./internal/budget/usecase/
```
