# Tests

## Purpose
Enforce consistent testing patterns, structure, and coverage.

## Scope
All `*_test.go` files in the project.

## Requirements

### Framework
- Must use `testify/suite` for use case and integration tests.
- Must use `testify/require` or suite assertion methods for expectations.
- Must use `testify/mock` for mock generation and assertions.

### Test Execution
```bash
go test ./...
go test ./... -cover
```

### Suite Structure
- Use case tests must use `suite.Suite` with `SetupTest` for per-test initialization.
- Must register suites via `TestXxxSuite(t *testing.T)` function.

### Table-Driven Tests
- Use case tests must use table-driven pattern with `scenarios` slice.
- Each scenario must define: `name`, `args`, `dependencies`, and `expect`.

### Independence
- Tests must not share mutable state.
- Each test must set up its own data.
- Must use `SetupTest` (not `SetupSuite`) for per-test reset.

### AAA Pattern
- Must follow Arrange, Act, Assert structure.
- Dependencies setup goes in the scenario definition (Arrange).
- Use case execution is the Act.
- `expect` function is the Assert.

### Mocks
- Must generate mocks using `mockery` in `infrastructure/repositories/mocks/`.
- Must use `EXPECT()` for mock expectations with `mock.AnythingOfType` for typed matchers.
- Must call `.Once()` or `.Times(n)` on mock expectations.

### Value Object and Entity Tests
- Must use standard `testing.T` with `require` for domain-level tests.
- No suite required for simple unit tests.

### Naming
- Test names must describe expected behavior.
- Must use descriptive names in English or Portuguese — match the existing codebase convention.

### Observability in Tests
- Must use `fake.NewProvider()` for observability in test suites.
- Must never use real observability providers in tests.

### Coverage
- All use case code must be covered by tests.
- All domain entities and value objects must be covered by tests.

## Examples

### Suite Setup
```go
type CreateCategoryUseCaseSuite struct {
    suite.Suite
    ctx  context.Context
    obs  observability.Observability
    repo *mocks.CategoryRepository
}

func TestCreateCategoryUseCaseSuite(t *testing.T) {
    suite.Run(t, new(CreateCategoryUseCaseSuite))
}

func (s *CreateCategoryUseCaseSuite) SetupTest() {
    s.obs = fake.NewProvider()
    s.ctx = context.Background()
    s.repo = mocks.NewCategoryRepository(s.T())
}
```

### Table-Driven Scenarios
```go
scenarios := []struct {
    name         string
    args         args
    dependencies dependencies
    expect       func(output *dtos.Output, err error)
}{
    {
        name: "should create category successfully",
        args: args{...},
        dependencies: dependencies{...},
        expect: func(output *dtos.Output, err error) {
            s.NoError(err)
            s.NotNil(output)
        },
    },
}

for _, scenario := range scenarios {
    s.Run(scenario.name, func() {
        uc := NewUseCase(s.obs, scenario.dependencies.repo)
        output, err := uc.Execute(s.ctx, scenario.args.input)
        scenario.expect(output, err)
    })
}
```

### Value Object Test
```go
func TestNewCategoryName(t *testing.T) {
    t.Run("should create valid name", func(t *testing.T) {
        name, err := vos.NewCategoryName("Transport")
        require.NoError(t, err)
        require.Equal(t, "Transport", name.String())
    })
}
```

### Test Naming
```go
// Required
"should create category successfully"
"should return error when name is empty"

// Forbidden
"test category"
"works"
```

## Forbidden
- Tests that depend on execution order of other tests.
- Tests without assertions.
- Tests that test multiple unrelated behaviors.
- Real database or network calls in unit tests.
- Shared mutable state between test cases.
- Real observability providers in tests.
