# Code Standards

## Purpose
Enforce consistent coding style and naming conventions across the Go codebase.

## Scope
All `.go` files in the project.

## Requirements

### Language
- All source code must be written in English: variable names, functions, structs, interfaces, constants, and comments.

### Naming Conventions
- **camelCase**: local variables, function parameters, unexported fields.
- **PascalCase**: exported functions, methods, structs, interfaces, constants.
- **snake_case**: file names and directory names.
- Interface names must not use `I` prefix. Must use behavior-based names (e.g., `CategoryRepository`, `TokenValidator`).

### Naming Clarity
- Must avoid abbreviations.
- Must avoid names longer than 30 characters.
- Methods and functions must start with a verb: `Create`, `Find`, `Update`, `Remove`, `Validate`, `List`, `Check`.
- Boolean variables must read as assertions: `isActive`, `hasPermission`, `canDelete`.

### Constants and Magic Numbers
- Must declare named constants for all magic numbers.

### Parameters
- Functions must not accept more than 3 positional parameters. Must use a struct for more.

### Early Returns
- Must not nest more than two levels of `if/else`. Must use early returns.

### Flag Parameters
- Must not use boolean flags to switch function behavior. Must extract into separate functions.

### Size Limits
- Functions: max 50 lines.
- Files: max 300 lines. Must split into focused files when exceeded.

### Variable Declaration
- One variable per line.
- Must declare variables close to where they are used.

### Blank Lines
- Must not add blank lines inside functions and methods.

### Comments
- Must avoid comments. Code must be self-explanatory.
- Exception: godoc comments on exported types and functions.

## Examples

### Naming
```go
// Forbidden
usrNm := "John"
userNameFromDatabaseQueryResult := "John"

// Required
userName := "John"
```

### Constants
```go
const (
    MinimumAge      = 18
    OneHourInMs     = 60 * 60 * 1000
    MaxPageSize     = 100
    DefaultPageSize = 20
)
```

### Parameters
```go
// Forbidden
func CreateUser(name, email string, age int, address, phone string) {}

// Required
type CreateUserParams struct {
    Name    string
    Email   string
    Age     int
    Address string
    Phone   string
}

func CreateUser(params CreateUserParams) {}
```

### Early Returns
```go
func processPayment(user *User, amount Money) error {
    if user == nil {
        return ErrUserRequired
    }
    if !user.IsActive {
        return ErrUserInactive
    }
    if !amount.IsPositive() {
        return ErrInvalidAmount
    }
    return completePayment(user, amount)
}
```

### Flag Parameters
```go
// Forbidden
func GetUser(id string, includeOrders bool) {}

// Required
func GetUser(id string) {}
func GetUserWithOrders(id string) {}
```

## Forbidden
- Portuguese variable or function names.
- Magic numbers without named constants.
- Nested conditionals deeper than two levels.
- Functions with more than 3 positional parameters.
- Boolean flag parameters.
- Multiple variable declarations on a single line.
- Comments that restate what the code does.
