# Box Project ´

These guidelines apply to all work in this repository.

## Tooling
- Makefile contains target for developing and testing.
- box command is a wrapper around make but is tied to the project and can be issued from anywhere
- box self-test is the command you start and end tasks with.

## Overall Objectives
- Provide an emulated environment where code can be tested exactly as it would run in production.
- Initially allowing for configuration changes, but ultimately treat code and configuration as a single immutable unit (container).
- Making emulation so well so that new features can be tested without having to deploy the project. 
- Making it possible to run and test compositions of services / frontends / backends.
- Local first development. Run and debug code using a localhost development environment.
- Only dockerize when necessary.
- Identifying weaknesses or issues before they go to production.

## Defeats The Purpose
- Use localhost/127.0.0.1 urls instead of DNS names.
- Weaken ssl security by ignoring or disabling ssl errors
- Solve issue by using something that would not work in production.
- Using 'hardening' or 'robust' configuration or features that can hide weaknesses or issues.

## Remember!
- There is a network inside docker and a network outside docker. DNS names resolve differently.
- To verify behavior inside docker use:
  - browser: https://browser.web.com
  - mcp: https://tools.web.internal/sse
- This is a development, mock and testing environment:
  - Do not apply security measures for other reasons than emulating production. Make it supereasy to bypass.
  - Do not use persistence, restarts should always start from same clean state.
- FAIL FAST AND LOUD; it is the only way to find weaknesses and issues.


## GO
### Writing Unit Tests

Unit tests MUST be fast, deterministic, and isolated.

### 1. Principles

- **Refactor over mocking**: Testability of code is a quality measure per see. Therefore mocking in unit tests is bad code smell and missing the point.
- **One concept per test**: Each test function SHOULD focus on a single behavior or invariant.
- **Independence**: Tests MUST NOT depend on each other or have a specific execution order.
- **Fast execution**: Unit tests MUST NOT perform I/O (filesystem, network, database). Use interfaces and mocks if necessary.
- **Clarity over cleverness**: Test code should be as readable as production code.

### 3. Mandatory Verification Rules

You MUST follow these rules:

1.  **Assert Field Matching**: For every constructor or builder, tests MUST verify that each field provided as input matches the corresponding getter output.
2.  **Test Panic Branches**: `Must*` constructors MUST have at least one test that exercises the panic branch using `defer`/`recover`.
3.  **No Silent Failures**: Every test MUST make at least one assertion using the testing framework (`t.Error`, `t.Fatal`, `require.*`, `assert.*`).
4.  **No Discarding Returns**: Tests MUST NOT discard return values with `_` unless explicitly commented as irrelevant.

### Example: Testing a Panic Branch

```go
func TestMustLeadCreatedAt_PanicsOnZero(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("The code did not panic")
        }
    }()

    // This should panic
    MustLeadCreatedAt(time.Time{})
}
```

### 4. Best Practices

- **Use `testify/require` for prerequisites**: Use `require` for checks that should stop the test immediately (e.g., `require.NoError(t, err)`).
- **Use `testify/assert` for values**: Use `assert` for checks that allow the test to continue (e.g., `assert.Equal(t, expected, actual)`).
- **Table-Driven Tests**: Use table-driven tests for complex logic with multiple input combinations.


