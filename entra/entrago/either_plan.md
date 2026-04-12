I can’t create or modify files in this environment (read-only). Instead, below is the complete content you can paste into a new file named `either_plan.md`.

### `either_plan.md`
```markdown
### Either Migration Plan (strict `domain.Error` everywhere)

#### Goal
- Any function that can fail MUST return `mo.Either[domain.Error, T]`.
- No fallible function may return `(T, error)` or `error`.
- All failures MUST be represented as `domain.Error` (value type), unified via `ErrorCode`.
- Tests SHOULD use `.MustRight()` / `.MustLeft()`.

#### Non-goals (for now)
- This plan starts with the `domain` package. Other packages (`config`, `app`, `http`) migrate after domain is stable.

#### Constraints (from `github.com/samber/mo` v1.16.0)
- `mo.Either[L, R]` methods are limited for cross-type pipelines.
- For pipelines that change the right type (`A -> Either[L, B]`), use free functions in `github.com/samber/mo/either`:
  - `either.MapRight` (pure mapping, changes right type)
  - `either.FlatMapRight` (bind/flatMap, changes right type)
  - `either.PipeN` (composition)

#### Canonical types
- Left type MUST be `domain.Error` (value), not `*domain.Error`.
- Sentinel errors MUST be values:
  - ✅ `var ErrTenantIDInvalid = domain.Error{...}`
  - ❌ `var ErrTenantIDInvalid = &domain.Error{...}`

#### Domain error identity
- `domain.Error` MUST contain an `ErrorCode`.
- Code comparison is the identity rule.
- If Go `errors.Is` is used anywhere, `domain.Error` SHOULD implement `Is(target error) bool` using `ErrorCode`.

#### Migration phases

##### Phase 0 — Inventory + acceptance gate
- Acceptance command: `make ci` (from `entra/entrago`).
  During Phase 0 to Phase
- Inventory all functions returning:
  - `(T, error)`
  - `error`
  - `*domain.Error`
  - mixed error types

##### Phase 1 — Normalize `domain.Error` to strict value type
- Convert all sentinel errors to `domain.Error` values.
- Ensure constructors and validators return `mo.Either[domain.Error, T]`.
- Ensure no package exports `*domain.Error`.
— Delete all mustTenantID like test helpers

##### Phase 2 — Convert `domain.New*` constructors
For each `NewXxx` that can fail:
- Change signature to:
  - `func NewXxx(...) mo.Either[domain.Error, Xxx]`
- Map any third-party/library errors immediately to a `domain.Error`.
- Start a leafs on domain objects with just one primitive value.  



Recommended minimal pattern (skip “empty” checks unless the domain truly needs distinct codes):

```go
func NewTenantID(raw string) mo.Either[domain.Error, TenantID] {
    id, err := uuid.Parse(raw)
    if err != nil {
        return mo.Left[domain.Error, TenantID](ErrTenantIDInvalid)
    }

    return mo.Right[domain.Error, TenantID](TenantID{value: id})
}
```

##### Phase 3 — Compose multiple failing steps using `mo/either`
Use `either.PipeN` and `either.FlatMapRight` / `either.MapRight`.

Example: building a tenant from raw input (cross-type pipeline) without `(T, error)`:

```go
import (
    "github.com/samber/mo"
    moeither "github.com/samber/mo/either"
)

func buildTenant(raw RawTenant) mo.Either[domain.Error, *domain.Tenant] {
    return moeither.Pipe3(
        mo.Right[domain.Error, RawTenant](raw),
        moeither.FlatMapRight(func(r RawTenant) mo.Either[domain.Error, domain.TenantID] {
            return domain.NewTenantID(r.TenantID)
        }),
        moeither.FlatMapRight(func(id domain.TenantID) mo.Either[domain.Error, *domain.Tenant] {
            // Continue with more `FlatMapRight` steps as needed,
            // or switch to a typed build-state if many values must be carried.
            return moeither.Pipe2(
                domain.NewTenantName(raw.Name),
                moeither.FlatMapRight(func(name domain.TenantName) mo.Either[domain.Error, domain.Tenant] {
                    return domain.NewTenant(id, name /* ... */)
                }),
            ).
            // Convert `domain.Tenant` value to `*domain.Tenant` pointer
            // (pointer is optional; keep value if you can).
            Match(
                func(e domain.Error) mo.Either[domain.Error, *domain.Tenant] { return mo.Left[domain.Error, *domain.Tenant](e) },
                func(t domain.Tenant) mo.Either[domain.Error, *domain.Tenant] { return mo.Right[domain.Error, *domain.Tenant](&t) },
            )
        }),
    )
}
```

Notes:
- Prefer returning value types (`domain.Tenant`) unless you have a strong reason to return pointers.
- If you need many intermediate values, use a typed build-state on the right.

##### Phase 4 — Tests
- Delete all`mustXxx(t, ...)` helpers 
  - Every test MUST assert returned values (not only “no error”).
  -Only use `ErrorCode` rather than full strings.

Example:

```go
func TestNewTenantID_Invalid(t *testing.T) {
    err := domain.NewTenantID("not-a-uuid").MustLeft()
    if err.Code != domain.ErrCodeTenantIDInvalid {
        t.Fatalf("code=%v", err.Code)
    }
}
```

#### Error mapping rules
- No `fmt.Errorf`/`errors.New` as business errors.
- External errors (IO/YAML/JWT/etc.) must be converted immediately to `domain.Error` with a stable code.
- Prefer codes that support stable HTTP mapping and stable tests.

#### Definition of Done (per phase)
- No `(T, error)` return values remain in the targeted package.
- No `error` is returned/propagated as a business failure.


#### Practical checklist for each PR
- [ ] Convert signatures to `mo.Either[domain.Error, T]`.
- [ ] Replace call sites with `either.FlatMapRight` / `either.MapRight` pipelines.
- [ ] Update tests to `.MustRight()` / `.MustLeft()` and assert `ErrorCode`.
- [ ] Remove dead code introduced by migration.

##
Final check
run make ci and start dealing with lint errors