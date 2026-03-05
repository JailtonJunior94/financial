# Repository

## Purpose
Enforce consistent repository implementation patterns for database access.

## Scope
All files under `internal/*/infrastructure/repositories/`.

## Requirements

### Structure
- Must be a private struct implementing a domain interface from `domain/interfaces/`.
- Constructor must accept `database.DBTX`, `observability.Observability`, and `*metrics.FinancialMetrics`.
- Constructor must return the domain interface type.

### Database Access
- Must use `database.DBTX` abstraction for all database operations.
- Must use parameterized queries with `$N` placeholders. Must never concatenate user input into SQL.
- Must use `PrepareContext` + `ExecContext` for write operations (insert, update).
- Must use `QueryContext` for multi-row reads and `QueryRowContext` for single-row reads.
- Must always close rows via `defer func() { _ = rows.Close() }()`.
- Must always close statements via `defer func() { _ = stmt.Close() }()`.
- Must check `rows.Err()` after iterating rows.

### Not Found Handling
- Must return `(nil, nil)` when a record is not found and absence is not an error.
- Must detect `sql.ErrNoRows` for single-row queries.

### Observability
- Every method must start a span: `{entity}_repository.{operation}`.
- Every method must capture `start := time.Now()` for metrics.
- Must log `query_started` (Debug) at method entry.
- Must log `query_completed` (Debug) on success.
- Must log `query_failed` (Error) on failure with `observability.Error(err)`.
- Must call `span.RecordError(err)` before returning errors.
- Must call `fm.RecordRepositoryQuery` on success.
- Must call `fm.RecordRepositoryFailure` on failure.

### Logging Fields
- Every log entry must include: `operation`, `layer` (`"repository"`), `entity`, `user_id`.
- Must include relevant entity IDs (e.g., `category_id`, `card_id`) when available.

## Examples

### Constructor
```go
type categoryRepository struct {
    db   database.DBTX
    o11y observability.Observability
    fm   *metrics.FinancialMetrics
}

func NewCategoryRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.CategoryRepository {
    return &categoryRepository{db: db, o11y: o11y, fm: fm}
}
```

### Method Pattern
```go
func (r *categoryRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error) {
    start := time.Now()
    ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list")
    defer span.End()
    r.o11y.Logger().Debug(ctx, "query_started",
        observability.String("operation", "list"),
        observability.String("layer", "repository"),
        observability.String("entity", "category"),
        observability.String("user_id", userID.String()),
    )
    query := `SELECT id, name FROM categories WHERE user_id = $1 AND deleted_at IS NULL`
    rows, err := r.db.QueryContext(ctx, query, userID.String())
    if err != nil {
        span.RecordError(err)
        r.o11y.Logger().Error(ctx, "query_failed",
            observability.String("operation", "list"),
            observability.String("layer", "repository"),
            observability.String("entity", "category"),
            observability.String("user_id", userID.String()),
            observability.Error(err),
        )
        r.fm.RecordRepositoryFailure(ctx, "list", "category", "infra", time.Since(start))
        return nil, err
    }
    defer func() { _ = rows.Close() }()
    // ... scan rows ...
    r.fm.RecordRepositoryQuery(ctx, "list", "category", time.Since(start))
    return categories, nil
}
```

## Forbidden
- Business logic in repositories.
- Returning domain errors from repositories (must return raw infrastructure errors).
- SQL string concatenation with user input.
- Ignoring `rows.Err()` after iteration.
- Missing observability (spans, logs, metrics) in any repository method.
- Direct use of `*sql.DB` instead of `database.DBTX`.
