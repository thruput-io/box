# Claims domain encapsulation plan (strict typing)

## Goals

- Claim keys MUST NOT exist as strings anywhere except inside `domain/types.go`.
- Claims MUST NOT be represented as `map[string]any`, `jwt.MapClaims`, or `map[string]string` anywhere except in `Claims.ToJWT()`.
- Claims MUST be validated/parsed at ingress (the test-token process) via an allowlist, and from that point onward MUST be domain objects.
- The only allowed egress back to JWT is `Claims.ToJWT()`, and it SHOULD be invoked exactly once at the boundary where `jwt.NewWithClaims(...)` is called.
- All fallible functions MUST return `mo.Either[domain.Error, T]` (strict `domain.Error` value type).

## Non-goals

- This document does not implement the refactor; it defines the strict end-state API and a phased migration plan.

## Constraints (from `github.com/samber/mo@v1.16.0`)

- Cross-type pipelines MUST use `github.com/samber/mo/either` free functions:
  - `either.MapRight` / `either.FlatMapRight`
  - `either.PipeN`

## Public domain API (target)

The public surface MUST remain opaque: neither claim keys nor claim values are to be readable as strings from outside the domain.

### Opaque allowlisted key

- `ValidClaim(rawKey string) mo.Option[ValidClaimKey]`
  - Returns `None` if `rawKey` is not allowlisted.
  - `ValidClaimKey` is exported but opaque (cannot be constructed by callers).

`ValidClaimKey` MUST provide fallible constructors:

- `func (k ValidClaimKey) NewClaimFromAny(v any) mo.Either[domain.Error, Claim]`
- `func (k ValidClaimKey) NewClaimArrayFromAny(v any) mo.Either[domain.Error, Claim]`

Rationale:
- Ingress parsing is where raw JSON types exist (`string`, `float64`, `[]any`, `json.Number`, etc.).
- After parsing, the system uses `Claim`/`Claims` only.

### Claims

- `type Claim` MUST be an exported opaque value type.
- `type Claims` MUST be an exported immutable value object.

`Claims` MUST support exactly one mutation-like operation, returning a new instance:

- `func (c Claims) Append(cl Claim) mo.Either[domain.Error, Claims]`

Append MUST:
- be immutable (copy-on-write of internal collections)
- short-circuit on error
- handle “same key appended multiple times” by promoting scalar → array and appending
- reject type-mismatches deterministically as `domain.Error` (see **Errors**)

### Test-only assertions (strict typed)

Tests MUST ONLY use the following assertion helpers:

- `ValidClaim(...)`
- `Claim.HasValue(...)`
- `Claims.ContainsValue(...)`

To keep assertions strict and prevent primitives:

- `HasValue` / `ContainsValue` MUST accept an exported interface with an unexported method, so only domain types can be used:

```go
type ClaimComparable interface {
	claimComparable()
}
```

Then:

- `func (c Claim) HasValue(v ClaimComparable) bool`
- `func (c Claims) ContainsValue(k ValidClaimKey, v ClaimComparable) bool` (if needed)

Domain primitive wrappers that are used as claim values MUST implement `ClaimComparable` (e.g., `TenantID`, `UserID`, `ClientID`, `Issuer`, `Nonce`, …).

For numeric claims, the domain MUST introduce a dedicated wrapper (e.g., `JwtNumericDateSeconds`) implementing `ClaimComparable`, so tests never pass raw `int64`.

## Internal domain model (strictest typed version)

### Claim key: no strings outside `types.go`

The strictest internal representation SHOULD avoid `struct{ key string }` entirely and use an unexported enum:

```go
type claimKey uint8
```

`types.go` MUST contain the only mapping between:

- the external string key (e.g. "sub")
- the internal `claimKey` value
- the claim specification (value kind + cardinality)

This makes illegal keys unrepresentable beyond the ingress validation.

### Claim value: type-safe union (string | int64)

`claimValue` MUST be a closed, non-interface tagged union to prevent nil/zero-interface states:

```go
type claimValueKind uint8

const (
	claimValueString claimValueKind = iota + 1
	claimValueInt64
)

type claimValue struct {
	kind claimValueKind
	s    NonEmptyString
	n    int64
}
```

Construction MUST be centralized and strict:

- `newClaimStringValue(NonEmptyString) claimValue`
- `newClaimInt64Value(JwtNumericDateSeconds) claimValue` (or equivalent)

Equality MUST be type-safe (string never equals int64).

### Scalar/array payload with a common base

Scalar and array claims MUST share one internal representation so the rest of the domain does not branch on “Claim vs ClaimArray”.

Strictest model:

```go
type claimCardinality uint8

const (
	claimScalar claimCardinality = iota + 1
	claimArray
)

type claimPayload struct {
	card claimCardinality
	one  claimValue
	many NonEmptyArray[claimValue]
}
```

Rules:

- `claimPayload` MUST always contain exactly one active variant.
- Promotion scalar → array MUST be done only inside `Claims.Append`.
- Arrays MUST be immutable (copy on append) and MUST stay non-empty via `NonEmptyArray`.

### Claims storage and immutability

`Claims` MUST be immutable and MUST have deterministic behavior.

Recommended strict internal form:

- `map[claimKey]claimPayload` for fast lookup
- `[]claimKey` insertion order to keep deterministic `ToJWT()` emission order (no reliance on Go map iteration)

On `Append`:

- clone the map and order slice (copy-on-write)
- update/insert one entry
- return the new `Claims` instance

## Ingress parsing rules

Ingress parsing MUST fail fast and return `domain.Error` (not `error`).

### Allowlist and spec

`types.go` MUST define a claim spec table:

- which keys are allowed
- whether they are scalar or array
- whether the value kind is string or int64

Example kinds (not exhaustive):

- string scalar: `sub`, `nonce`, `uti`, `tid`, `ver`, `oid`, `azp`, `appid`, `iss`, `aud`, `typ`, `name`, `preferred_username`, `email`, `unique_name`
- string array: `roles`, `groups`, `scp`
- int64 scalar: `exp`, `iat`, `nbf`

If a claim is not in the allowlist, parsing MUST return `ErrClaimKeyNotAllowed`.

### Numeric conversion (strict)

For numeric claims (`exp`, `iat`, `nbf`):

- Accept `json.Number` or `float64` (typical `encoding/json` output).
- Convert to `int64` ONLY if the number is integral and within int64 range.
- Reject fractional values (e.g. `123.4`) deterministically.

### Array conversion (strict)

- Accept `[]any`.
- Validate every element against the claim spec (string vs int64).
- Reject empty arrays using the unified `ErrNonEmptyArrayEmpty`.

## Egress (`ToJWT`) rules

- `Claims.ToJWT()` MUST emit `jwt.MapClaims` with correct Go types:
  - string scalar → `string`
  - int64 scalar → `int64`
  - string array → `[]string`
  - int64 array (if ever supported) → `[]int64`
- `ToJWT()` MUST be the only domain method that converts claims to key strings.

## Errors (domain.Error)

All failures in this subsystem MUST be `domain.Error` (value type). Proposed error codes:

- `ErrClaimKeyNotAllowed`
- `ErrClaimTypeMismatch` (value kind mismatch for a given key)
- `ErrClaimCardinalityMismatch` (scalar vs array mismatch when parsing raw input)
- `ErrClaimNumericNotIntegral`
- `ErrClaimNumericOutOfRange`

Note: array emptiness MUST reuse the already-existing unified domain error (e.g., `ErrNonEmptyArrayEmpty`).

## Refactor plan (phases)

### Phase 1 — Lock down key strings

- Move/convert exported `SubClaim`, `TidClaim`, … to internal-only key specs.
- Change `ValidClaim` to return `mo.Option[ValidClaimKey]` (opaque wrapper).
- Ensure `ValidClaim` returns `None` when the key is missing (no default/zero-key acceptance).

### Phase 2 — Typed values + payload base

- Introduce `claimValue` tagged union (string|int64) + `claimPayload` (scalar|array).
- Implement strict `Claims.Append` (immutable, scalar→array promotion, deterministic).

### Phase 3 — Ingress parsing uses domain

- In the test-token flow:
  1. parse raw JWT claims as `map[string]any`
  2. for each key:
     - `ValidClaim(key)`
     - `NewClaimFromAny` / `NewClaimArrayFromAny`
     - `Claims.Append(...)`
  3. on first failure: map `domain.Error` → HTTP response

### Phase 4 — Remove string-claim usage in app/http/tests

- Remove claim-key string constants from `app`.
- Replace any direct map indexing (`claims["sub"]`) with domain assertions and/or domain-derived claim building.
- Update tests to only use `ValidClaim`, `HasValue`, `ContainsValue` and to assert via domain types (no primitives).

### Phase 5 — Egress boundary only

- Ensure the only call site that produces `jwt.MapClaims` is `Claims.ToJWT()`.
- Ensure `Claims.ToJWT()` is called exactly once before `jwt.NewWithClaims(...)`.

## Acceptance criteria (Definition of Done)

- No claim-key strings exist outside `domain/types.go`.
- No claim values are exposed via exported fields or getters.
- All fallible claim operations return `mo.Either[domain.Error, T]`.
- Numeric claims round-trip as numeric values (`int64`) through `ToJWT()`.
- `Claims.Append` is immutable and deterministic.

## Verification (when implementation starts)

- `gofmt -w ./...`
- `go test ./...`
- `make ci` (from `entra/entrago`) without muting/skipping tests
