## GO

Domain objects should be created directly at the boundaries of the application.
 - When parsing configuration
 - When parsing incoming query parameters or request bodies  

Domain are the only allowed objects to be used in the application in any values returned from methods or used as arguments

Domain objects should be Opaque, meaning all fields should be private and never read directly.
Domain objects should be immutable. 
Domain objects can be combined, transformed merged in any way.
Domain object always have a semantic meaning.
Thereofre a domain object can never be a primitive type or a generic value object type.

IMPORTANT: Primitives are not allowed to be used in the application.
Exceptions are:
- When parsing configuration
- When parsing incoming query parameters or request bodies
- When used as output in a web page or API response, however, if possible should use other representations than strings.

Unit tests should expected outcomes via domain objects.







### Writing Unit Tests

Unit tests MUST deterministic, and isolated.

### Workflow
1. Check guidelines
2. Check git state 
3. Run make all – everything should be green


### 1. Principles

- **Refactor over mocking**: Testability of code is a quality measure per see. Therefore mocking in unit tests is bad code smell and missing the point.
- **One concept per test**: Each test function SHOULD focus on a single behavior or invariant.
- **Independence**: Tests MUST NOT depend on each other or have a specific execution order.
- **Clarity over cleverness**: Test code should be as readable as production code.

### 3. Mandatory Verification Rules

You MUST follow these rules:

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


