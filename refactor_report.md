# Refactor Report — R-ARCH-001: Defer Close with Observability

- **Mode:** execution
- **Date:** 2026-03-10
- **Rule:** R-ARCH-001 (hard)
- **Pattern replaced:** `defer func() { _ = resource.Close() }()` → full observability-aware defer

---

## Hotspots (Files Altered)

| # | File |
|---|------|
| 1 | `internal/transaction/infrastructure/repositories/transaction_repository.go` |
| 2 | `internal/invoice/infrastructure/repositories/invoice_repository.go` |
| 3 | `internal/budget/infrastructure/repositories/budget_repository.go` |
| 4 | `internal/payment_method/infrastructure/repositories/payment_method_repository.go` |
| 5 | `internal/user/infrastructure/repositories/user_repository.go` |
| 6 | `internal/user/infrastructure/repositories/user_repository_write.go` |
| 7 | `internal/category/infrastructure/repositories/category_repository.go` |
| 8 | `internal/category/infrastructure/repositories/subcategory_repository.go` |
| 9 | `internal/category/infrastructure/adapters/category_provider_adapter.go` |
| 10 | `internal/category/infrastructure/adapters/category_provider_adapter_test.go` |

---

## Changes (Detail per File)

### 1. `transaction_repository.go`
- **`Save`** — `defer func() { _ = stmt.Close() }()` → span + logger error on close failure for `stmt`
- **`FindByInstallmentGroup`** — `defer func() { _ = rows.Close() }()` → span + logger error on close failure for `rows`
- **`Update`** — `defer func() { _ = stmt.Close() }()` → span + logger error on close failure for `stmt`
- **`ListPaginated`** — `defer func() { _ = rows.Close() }()` → span + logger error on close failure for `rows`

### 2. `invoice_repository.go`
- **`FindByUserAndMonth`** — `rows.Close()` → span + logger (span available)
- **`ListByUserAndMonthPaginated`** — `rows.Close()` → span + logger (span available)
- **`FindByCard`** — `rows.Close()` → span + logger (span available)
- **`ListByCard`** — `rows.Close()` → span + logger (span available)
- **`FindItemsByPurchaseOrigin`** — `rows.Close()` → span + logger (span available)
- **`findItemsByInvoiceIDs`** (private helper) — `rows.Close()` → logger only (no span in scope)
- **`findItemsByInvoiceID`** (private helper) — `rows.Close()` → logger only (no span in scope)

### 3. `budget_repository.go`
- **`ListPaginated`** — `rows.Close()` → span + logger (span available)
- **`findItemsByBudgetIDs`** (private helper) — `rows.Close()` → logger only (no span in scope)
- **`findItemsByBudgetID`** (private helper) — `rows.Close()` → logger only (no span in scope)

### 4. `payment_method_repository.go`
- **`ListPaginated`** — `rows.Close()` → span + logger (span available)

### 5. `user_repository.go`
- **`Insert`** — `stmt.Close()` → span + logger (span available)

### 6. `user_repository_write.go`
- **`scanUsers`** (private helper) — `rows.Close()` → logger only (no span in scope)
- **`Update`** — `stmt.Close()` → span + logger (span available)
- **`SoftDelete`** — `stmt.Close()` → span + logger (span available)

### 7. `category_repository.go`
- **`ListPaginated`** — `rows.Close()` → span + logger (span available)

### 8. `subcategory_repository.go`
- **`FindByCategoryID`** — `rows.Close()` → span + logger (span available)
- **`ListPaginated`** — `rows.Close()` → span + logger (span available)

### 9. `category_provider_adapter.go`
- **`queryFoundIDs`** (private helper) — `rows.Close()` → logger only using `a.o11y` (no span in scope, receiver is `a`)

### 10. `category_provider_adapter_test.go`
- 5 test methods — `db.Close()` (sqlmock `*sql.DB`) → logger only using `s.obs.Logger()` (test-level, no span)
- Added import: `"github.com/JailtonJunior94/devkit-go/pkg/observability"` for `observability.Error()`

---

## Pattern Applied

**When span is available in scope:**
```go
defer func() {
    if closeErr := resource.Close(); closeErr != nil {
        span.RecordError(closeErr)
        r.o11y.Logger().Error(ctx, "{MethodName}: failed to close {resource}",
            observability.Error(closeErr),
        )
    }
}()
```

**When no span is available (private helpers):**
```go
defer func() {
    if closeErr := resource.Close(); closeErr != nil {
        r.o11y.Logger().Error(ctx, "{MethodName}: failed to close {resource}",
            observability.Error(closeErr),
        )
    }
}()
```

---

## Validation

| Gate | Verdict | Evidence |
|------|---------|----------|
| `go build ./...` | APPROVED | No output — zero compilation errors |
| `make test` | APPROVED | All test suites pass; total coverage 24.2% (unchanged from baseline) |

---

## Follow-up Refactoring Session (2026-03-11)

### Additional violations found and corrected

**File:** `internal/payment_method/infrastructure/repositories/payment_method_repository.go`

| Method | Violation | Fix |
|--------|-----------|-----|
| `List` | `defer rows.Close()` missing logger in o11y; `rows.Err()` not checked after iteration | Added `r.o11y.Logger().Error(...)` to close defer; added `rows.Err()` check |
| `Save` | `defer stmt.Close()` missing logger in o11y | Added `r.o11y.Logger().Error(...)` to close defer |
| `Update` | `defer stmt.Close()` missing logger in o11y | Added `r.o11y.Logger().Error(...)` to close defer |

### Files confirmed compliant (no changes needed)

- `invoice_repository.go` — all rows/stmt closes fully observability-aware
- `budget_repository.go` — all rows/stmt closes fully observability-aware
- `transaction_repository.go` — all rows/stmt closes fully observability-aware
- `user_repository.go` / `user_repository_write.go` — all closes fully observability-aware
- `category_repository.go` / `subcategory_repository.go` — all closes fully observability-aware
- All `domain/errormappings.go` stub files — empty package stubs, no violations
- All root-level `internal/{module}/errormappings.go` files — correctly wired via `module.go`

### Gate results (follow-up session)

| Gate | Verdict | Evidence |
|------|---------|----------|
| `make test` | APPROVED | 32 packages, all `ok`, no `FAIL` lines; total coverage 24.2% |

---

## Assumptions

- Private helper methods (`findItemsByInvoiceIDs`, `findItemsByInvoiceID`, `findItemsByBudgetIDs`, `findItemsByBudgetID`, `queryFoundIDs`, `scanUsers`) do not start their own spans; span logging was omitted per the rule's explicit directive.
- Test-level `db.Close()` in `category_provider_adapter_test.go` uses the test suite's `s.obs` (fake provider) as the observability source, which satisfies the rule without polluting test assertions.
- Empty `domain/errormappings.go` stub files are valid Go (package-only files); they were retained as-is since they compile without issue and may serve as future extension points.

---

## Final Cleanup (2026-03-11)

### Empty stubs removed

The following 7 files were reduced to `package domain` stubs after their content was migrated to module-root `errormappings.go`. They were deleted:

- `internal/budget/domain/errormappings.go`
- `internal/card/domain/errormappings.go`
- `internal/category/domain/errormappings.go`
- `internal/invoice/domain/errormappings.go`
- `internal/payment_method/domain/errormappings.go`
- `internal/transaction/domain/errormappings.go`
- `internal/user/domain/errormappings.go`

HTTP status mappings now live at `internal/{module}/errormappings.go` (module root), correctly outside the domain layer. `module.go` files in each bounded context already called `ErrorMappings()` from this location.

| Gate | Verdict | Evidence |
|------|---------|----------|
| `go build ./...` | APPROVED | Zero compilation errors |
| `make test` | APPROVED | 32/32 packages `ok`, coverage 24.2% |

---

## Residual Risk

**Low** — All changes are purely structural. No business logic, query, or data contract was altered.
