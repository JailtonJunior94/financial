# Error Handling

## Purpose
Enforce consistent error definition, propagation, and HTTP mapping across all layers.

## Scope
All `.go` files involving error creation, wrapping, or handling.

## Requirements

### Sentinel Errors
- Shared domain errors must be defined in `pkg/custom_errors/errors.go` as `var Err* = errors.New(...)`.
- Module-specific domain errors must be defined in `internal/{module}/domain/errors.go`.
- Error variable names must start with `Err` prefix: `ErrUserNotFound`, `ErrBudgetInvalidTotal`.
- Error messages must be lowercase, concise, and human-readable.

### Error Wrapping
- Must use `fmt.Errorf` with `%w` verb for wrapping errors.
- Wrapped errors must add context about the current operation.
- Must preserve the original error chain for `errors.Is` and `errors.As` checks.

### Error Propagation
- Domain layer: must return sentinel errors or wrapped sentinel errors.
- Use case layer: must propagate domain errors without modification. Must wrap infrastructure errors with context.
- Handler layer: must delegate all error responses to `errorHandler.HandleError(w, r, err)`.
- Repository layer: must return raw infrastructure errors. Must not translate to domain errors.

### Error-to-HTTP Mapping
- Must use `httperrors.ErrorMapper` to map errors to HTTP status codes.
- Base mappings must be registered in `httperrors.NewErrorMapper()`.
- Module-specific mappings must be passed via `NewErrorMapper(extra...)` parameter.
- Mapping priority: direct match -> `errors.Is` match -> JSON error detection -> validation heuristic -> 500.

### Error Response Format
- All error responses must use `httperrors.ProblemDetail` (RFC 7807).
- Must never expose internal error messages to clients.

### Error Checking
- Must use `errors.Is` for sentinel error comparison.
- Must use `errors.As` for typed error extraction.
- Must never compare errors with `==` operator.

## Examples

### Sentinel Error Definition
```go
// pkg/custom_errors/errors.go
var (
    ErrUserNotFound   = errors.New("user not found")
    ErrCardNotFound   = errors.New("card not found")
    ErrInvalidEmail   = errors.New("invalid email format")
)
```

### Module-Specific Errors
```go
// internal/invoice/domain/errors.go
var (
    ErrInvoiceClosed    = errors.New("invoice is closed")
    ErrDuplicatePurchase = errors.New("duplicate purchase in invoice")
)
```

### Error Wrapping
```go
return fmt.Errorf("failed to create category: %w", customErrors.ErrNameIsRequired)
```

### Error Mapping Registration
```go
errorMapper := httperrors.NewErrorMapper(map[error]httperrors.ErrorMapping{
    domain.ErrInvoiceClosed: {Status: http.StatusUnprocessableEntity, Message: "Invoice is closed"},
})
```

## Forbidden
- Silently swallowing errors.
- Using `==` for error comparison instead of `errors.Is`.
- Exposing internal error details in HTTP responses.
- Defining domain errors in infrastructure or application layers.
- Logging errors without also returning or handling them.
- Using `panic` for recoverable errors.
