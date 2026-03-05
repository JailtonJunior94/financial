# Go Architecture

## Purpose
Enforce Go idiomatic design, Clean Architecture, and DDD patterns across the codebase.

## Scope
All Go source code in the project.

## Requirements

### Project Structure
Each bounded context lives under `internal/{module}/` with this layout:

```
internal/{module}/
├── domain/
│   ├── entities/       # Aggregates and entities
│   ├── vos/            # Value objects
│   ├── factories/      # Entity creation with validation
│   ├── interfaces/     # Repository and service contracts
│   ├── strategies/     # Domain strategies (optional)
│   └── errors.go       # Domain-specific sentinel errors
├── application/
│   ├── dtos/           # Input/output data transfer objects
│   └── usecase/        # Use case implementations and interfaces
├── infrastructure/
│   ├── http/           # Handlers and routers
│   ├── repositories/   # Repository implementations
│   │   └── mocks/      # Generated mocks (mockery)
│   └── adapters/       # External service adapters
└── module.go           # Module wiring (composition root)
```

Shared code lives under `pkg/`:
```
pkg/
├── api/                # HTTP utilities, middlewares, error handling
│   ├── httperrors/     # ProblemDetail, ErrorHandler, ErrorMapper
│   ├── middlewares/    # Auth, metrics, ownership middlewares
│   └── http/           # Shared HTTP resource utilities
├── auth/               # Token validation and generation (JwtAdapter)
├── constants/          # Shared constants (currency, etc.)
├── custom_errors/      # Shared sentinel errors and CustomError type
├── database/           # Database abstractions (DBTX)
├── domain/
│   ├── vos/            # Shared value objects (ReferenceMonth, etc.)
│   └── interfaces/     # Cross-module domain interfaces
├── helpers/            # Utility functions (time helpers, etc.)
├── jobs/               # Background job definitions
├── lifecycle/          # Service lifecycle management
├── messaging/          # Event/message abstractions (consumer, handler)
├── money/              # Monetary rounding utilities
├── observability/
│   └── metrics/        # Module-specific metric structs (FinancialMetrics)
├── outbox/             # Transactional outbox pattern implementation
├── pagination/         # Cursor-based pagination
├── scheduler/          # Job scheduling
└── validation/         # Input validation utilities
```

### Layer Responsibilities

#### Domain Layer
- Contains entities, value objects, factories, interfaces, and domain errors.
- Must not import from `application`, `infrastructure`, or `pkg/api`.
- Entities must validate their own invariants.
- Value objects must be immutable and self-validating.
- Factories must bridge raw input to domain entities via value objects.
- Repository interfaces must be defined here, implemented in infrastructure.

#### Application Layer (Use Cases)
- Orchestrates domain logic. Must not contain business rules.
- Each use case must define a public interface and a private implementation struct.
- Constructor must return the interface type.
- Must receive repository interfaces and observability via constructor injection.
- Must map between DTOs and domain entities via factories.

#### Infrastructure Layer
- Implements domain interfaces (repositories, adapters).
- Handlers belong here as the transport layer.
- Must not contain business logic.

#### Module (Composition Root)
- Wires all dependencies for a bounded context.
- Creates repositories, use cases, handlers, and routers.
- Must be exported as a struct with the router field.
- Must accept `database.DBTX`, `observability.Observability`, and auth interfaces.

### Dependency Rule
- Dependencies must point inward: infrastructure -> application -> domain.
- Domain must never depend on outer layers.
- Application must depend only on domain interfaces.
- Infrastructure must implement domain interfaces.

### Value Objects
- Must be self-validating structs.
- Constructor (`New*`) must return `(VO, error)`.
- Must expose a `String()` method or `Value` field for reading.

### Entities
- Constructor (`New*`) must receive only value objects and validated types.
- Mutation methods must validate invariants and return `error` when invalid.
- Must use `Delete()` for soft-delete pattern (sets `DeletedAt`).

### Factories
- Must bridge raw input (strings, ints) and domain entities.
- Must parse and validate all inputs, create value objects, then call entity constructor.
- Must generate UUIDs for new entities.

### Context
- All public methods must accept `context.Context` as first parameter.
- Must pass context through the entire call chain.

### Interfaces
- Must be defined in the consumer package (domain), not the implementation package.
- Must be small and focused (Interface Segregation Principle).

## Examples

### Use Case Definition
```go
type (
    CreateCategoryUseCase interface {
        Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
    }
    createCategoryUseCase struct {
        o11y       observability.Observability
        repository interfaces.CategoryRepository
    }
)

func NewCreateCategoryUseCase(
    o11y observability.Observability,
    repository interfaces.CategoryRepository,
) CreateCategoryUseCase {
    return &createCategoryUseCase{o11y: o11y, repository: repository}
}
```

### Module Wiring
```go
type CategoryModule struct {
    CategoryRouter *http.CategoryRouter
}

func NewCategoryModule(db database.DBTX, o11y observability.Observability, tokenValidator auth.TokenValidator) CategoryModule {
    errorHandler := httperrors.NewErrorHandler(o11y)
    authMiddleware := middlewares.NewAuthorization(tokenValidator, o11y, errorHandler)
    financialMetrics := metrics.NewFinancialMetrics(o11y)
    categoryRepository := repositories.NewCategoryRepository(db, o11y, financialMetrics)
    // ... wire use cases, handlers, router
    return CategoryModule{CategoryRouter: categoryRouter}
}
```

### Value Object
```go
type CategoryName struct {
    Value *string
    Valid bool
}

func NewCategoryName(name string) (CategoryName, error) {
    trimmed := strings.TrimSpace(name)
    if len(trimmed) == 0 {
        return CategoryName{}, fmt.Errorf("invalid category name: %w", customErrors.ErrNameIsRequired)
    }
    return CategoryName{Value: &trimmed, Valid: true}, nil
}
```

## Forbidden
- Business logic in handlers or repositories.
- Domain layer importing from application or infrastructure.
- Concrete types in use case constructors (must use interfaces).
- Global mutable state.
- `init()` functions for business logic.
- Circular dependencies between modules.
- Raw SQL in use cases.
