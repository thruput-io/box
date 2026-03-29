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


