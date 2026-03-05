# HTTP Layer

## Purpose
Enforce consistent HTTP routing, handler design, and API conventions.

## Scope
All files under `internal/*/infrastructure/http/` and `pkg/api/`.

## Requirements

### Framework
- Must use `go-chi/chi` for HTTP routing.
- Must use `net/http` standard library for handlers (`http.HandlerFunc` signature).

### REST Resource Naming
- Resources must be in English, plural, and kebab-case.
- Resource depth must not exceed 3 levels.

### Mutations and Actions
- Specific actions must use POST with a verb in the path.
- PUT is only for full resource replacement.

### Router Structure
- Each module must have a dedicated `Router` struct with a `Register(chi.Router)` method.
- Authentication middleware must be applied via `router.Group`.

### Handler Structure
- Handlers must receive use cases via constructor injection.
- Handler methods must follow this execution order:
  1. Start span.
  2. Extract user from context.
  3. Decode body (if applicable).
  4. Validate input.
  5. Call use case.
  6. Respond.
- Must use `responses.JSON(w, statusCode, data)` for all responses.
- Must use `errorHandler.HandleError(w, r, err)` for all error responses.
- Must include structured logging with `operation`, `layer`, `entity`, `correlation_id`, and `user_id`.

### Request/Response Format
- All payloads must be JSON.
- Error responses must follow RFC 7807 Problem Details via `httperrors.ProblemDetail`.

### HTTP Status Codes
| Code | Usage |
|------|-------|
| 200  | Successful query or update |
| 201  | Successful resource creation |
| 204  | Successful deletion (no body) |
| 400  | Malformed request or validation error |
| 401  | Missing or invalid authentication |
| 403  | Authenticated but insufficient permissions |
| 404  | Resource not found |
| 409  | Resource conflict (duplicate, invalid reference) |
| 422  | Business rule violation |
| 500  | Unexpected server error |

### Middlewares
- Must use middlewares for cross-cutting concerns: authentication, metrics, request ID, ownership.
- Authentication middleware must extract Bearer token and inject `AuthenticatedUser` into context.
- Ownership middleware must validate resource ownership via user ID from context.

### Swagger/Godoc
- All handler methods must have godoc annotations for `swaggo/swag` generation.

## Examples

### Resource Paths
```
// Required
GET    /api/v1/categories
GET    /api/v1/categories/{id}
POST   /api/v1/categories
PUT    /api/v1/categories/{id}
DELETE /api/v1/categories/{id}
GET    /api/v1/cards/{cardId}/invoices

// Forbidden
GET /api/v1/getCategories
GET /api/v1/category/{id}
GET /api/v1/categoria/{id}
GET /api/v1/channels/{id}/playlists/{id}/videos/{id}/comments
```

### Action Paths
```
POST /api/v1/users/{userId}/change-password
POST /api/v1/orders/{orderId}/cancel
```

### Router
```go
type CategoryRouter struct {
    handlers       *CategoryHandler
    authMiddleware middlewares.Authorization
}

func (r CategoryRouter) Register(router chi.Router) {
    router.Group(func(protected chi.Router) {
        protected.Use(r.authMiddleware.Authorization)
        protected.Get("/api/v1/categories", r.handlers.Find)
        protected.Post("/api/v1/categories", r.handlers.Create)
    })
}
```

## Forbidden
- Business logic in handlers.
- Direct database access from handlers.
- Handlers without observability (tracing and logging).
- Raw `http.Error()` calls — must use `errorHandler.HandleError`.
- Handlers that silently swallow errors.
