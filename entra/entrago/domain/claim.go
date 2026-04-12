package domain

import (
	"strings"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

// ClaimComparable is a sealed interface implemented only by domain types
// that can be used as claim values. The unexported method prevents external implementations.
type ClaimComparable interface {
	claimComparable()
	toClaimValue() claimValue
}

// ValidClaimKey is an opaque wrapper around an allowlisted claimKey.
// Can only be obtained via ValidClaim.
type ValidClaimKey struct {
	key claimKey
}

// Claim is an opaque value type holding an allowlisted key and its typed payload.
type Claim struct {
	key     claimKey
	payload claimPayload
}

func newStringClaim(k claimKey, v NonEmptyString) Claim {
	return Claim{key: k, payload: newScalarPayload(newClaimStringValue(v))}
}

func newInt64Claim(k claimKey, n int64) Claim {
	return Claim{key: k, payload: newScalarPayload(newClaimInt64Value(n))}
}

func newStringArrayClaim(k claimKey, values []claimValue) Claim {
	return Claim{key: k, payload: newArrayPayload(values)}
}

// HasValue reports whether this claim holds the given domain value.
func (c Claim) HasValue(v ClaimComparable) bool {
	return c.payload.contains(v.toClaimValue())
}

// HasStringValue reports whether this claim holds the given raw string value.
// Returns false for an empty string since all stored claim values are non-empty.
func (c Claim) HasStringValue(s string) bool {
	if s == "" {
		return false
	}

	return c.payload.contains(newClaimStringValue(NonEmptyString{value: s}))
}

// HasInt64Value reports whether this claim holds the given raw int64 value.
func (c Claim) HasInt64Value(n int64) bool {
	return c.payload.contains(newClaimInt64Value(n))
}

// NewStringClaim constructs a typed string Claim for this key.
// Returns ErrClaimTypeMismatch if the key does not accept string values.
// Returns ErrNonEmptyStringEmpty if s is empty.
func (k ValidClaimKey) NewStringClaim(s string) mo.Either[Error, Claim] {
	if k.key.valueKind != claimValueKindString {
		return mo.Left[Error, Claim](ErrClaimTypeMismatch)
	}

	nesE := NewNonEmptyString(s)

	nes, ok := nesE.Right()
	if !ok {
		return mo.Left[Error, Claim](nesE.MustLeft())
	}

	return mo.Right[Error, Claim](newStringClaim(k.key, nes))
}

// NewInt64Claim constructs a typed int64 Claim for this key.
// Returns ErrClaimTypeMismatch if the key does not accept int64 values.
func (k ValidClaimKey) NewInt64Claim(n int64) mo.Either[Error, Claim] {
	if k.key.valueKind != claimValueKindInt64 {
		return mo.Left[Error, Claim](ErrClaimTypeMismatch)
	}

	return mo.Right[Error, Claim](newInt64Claim(k.key, n))
}

// NewStringArrayClaim constructs a typed string-array Claim for this key.
// Returns ErrClaimTypeMismatch if the key does not accept string values.
// Returns ErrClaimShapeInvalid if the key does not allow arrays.
// Returns ErrNonEmptyArrayEmpty if values is empty.
// Returns ErrNonEmptyStringEmpty if any element is empty.
func (k ValidClaimKey) NewStringArrayClaim(values []string) mo.Either[Error, Claim] {
	if k.key.valueKind != claimValueKindString {
		return mo.Left[Error, Claim](ErrClaimTypeMismatch)
	}

	if !k.key.allowsArray {
		return mo.Left[Error, Claim](ErrClaimShapeInvalid)
	}

	if len(values) == 0 {
		return mo.Left[Error, Claim](ErrNonEmptyArrayEmpty)
	}

	cv := make([]claimValue, 0, len(values))

	for _, s := range values {
		nesE := NewNonEmptyString(s)

		nes, ok := nesE.Right()
		if !ok {
			return mo.Left[Error, Claim](nesE.MustLeft())
		}

		cv = append(cv, newClaimStringValue(nes))
	}

	return mo.Right[Error, Claim](newStringArrayClaim(k.key, cv))
}

// JwtNumericDateSeconds wraps an int64 Unix timestamp for constructing numeric date claims.
// Use HasInt64Value(n) in tests — no ClaimComparable implementation needed.
type JwtNumericDateSeconds struct {
	value int64
}

// NewJwtNumericDateSeconds constructs a JwtNumericDateSeconds from a raw int64.
func NewJwtNumericDateSeconds(n int64) JwtNumericDateSeconds {
	return JwtNumericDateSeconds{value: n}
}

// AsExpClaim returns a typed exp claim.
func (j JwtNumericDateSeconds) AsExpClaim() Claim {
	return newInt64Claim(expClaim, j.value)
}

// AsIatClaim returns a typed iat claim.
func (j JwtNumericDateSeconds) AsIatClaim() Claim {
	return newInt64Claim(iatClaim, j.value)
}

// AsNbfClaim returns a typed nbf claim.
func (j JwtNumericDateSeconds) AsNbfClaim() Claim {
	return newInt64Claim(nbfClaim, j.value)
}

// TokenUniqueID wraps a UUID for constructing uti/jti claims.
// Use HasStringValue(id.String()) in tests — no ClaimComparable implementation needed.
type TokenUniqueID struct {
	value NonEmptyString
}

// TokenUniqueIDFromUUID constructs a TokenUniqueID from a uuid.UUID (infallible).
func TokenUniqueIDFromUUID(id uuid.UUID) TokenUniqueID {
	return TokenUniqueID{value: AsString(id)}
}

// AsUtiClaim returns a typed uti claim.
func (t TokenUniqueID) AsUtiClaim() Claim {
	return newStringClaim(utiClaim, t.value)
}

// AsJtiClaim returns a typed jti claim.
func (t TokenUniqueID) AsJtiClaim() Claim {
	return newStringClaim(jtiClaim, t.value)
}

// RolesClaimFrom constructs a roles array claim from a non-empty array of RoleValues.
func RolesClaimFrom(roles NonEmptyArray[RoleValue]) Claim {
	items := roles.Items()
	values := make([]claimValue, len(items))
	for i, r := range items {
		values[i] = newClaimStringValue(r.value)
	}

	return newStringArrayClaim(rolesClaim, values)
}

// EmptyRolesClaim constructs a roles claim with an empty array payload (roles: []).
func EmptyRolesClaim() Claim {
	return Claim{key: rolesClaim, payload: newArrayPayload(nil)}
}

// GroupsClaimFrom constructs a groups array claim from a non-empty array of GroupIDs.
func GroupsClaimFrom(groups NonEmptyArray[GroupID]) Claim {
	items := groups.Items()
	values := make([]claimValue, len(items))
	for i, g := range items {
		values[i] = newClaimStringValue(AsString(g.value))
	}

	return newStringArrayClaim(groupsClaim, values)
}

// ScpClaimFrom constructs an scp claim as a space-separated scalar string from a non-empty array of ScopeValues.
// Azure AD tokens represent scp as a single space-delimited string, not a JSON array.
func ScpClaimFrom(scopes NonEmptyArray[ScopeValue]) Claim {
	items := scopes.Items()
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = s.value.value
	}

	return newStringClaim(scpClaim, NonEmptyString{value: strings.Join(parts, " ")})
}

// RolesClaimFromSlice constructs a roles claim from a plain slice of RoleValues.
// Returns None if the slice is empty.
func RolesClaimFromSlice(roles []RoleValue) mo.Option[Claim] {
	if len(roles) == 0 {
		return mo.None[Claim]()
	}

	return mo.Some(RolesClaimFrom(NewNonEmptyArray(roles...).MustRight()))
}

// GroupsClaimFromSlice constructs a groups claim from a plain slice of GroupIDs.
// Returns None if the slice is empty.
func GroupsClaimFromSlice(groups []GroupID) mo.Option[Claim] {
	if len(groups) == 0 {
		return mo.None[Claim]()
	}

	return mo.Some(GroupsClaimFrom(NewNonEmptyArray(groups...).MustRight()))
}

// ScpClaimFromScopeSlice constructs an scp claim as a space-separated scalar string from a plain slice of ScopeValues.
// Returns None if the slice is empty.
func ScpClaimFromScopeSlice(scopes []ScopeValue) mo.Option[Claim] {
	if len(scopes) == 0 {
		return mo.None[Claim]()
	}

	return mo.Some(ScpClaimFrom(NewNonEmptyArray(scopes...).MustRight()))
}

// ScpClaimFromRoleSlice constructs an scp claim as a space-separated scalar string using RoleValue strings.
// Used for Graph API tokens where roles are expressed as scp values.
// Returns None if the slice is empty.
func ScpClaimFromRoleSlice(roles []RoleValue) mo.Option[Claim] {
	if len(roles) == 0 {
		return mo.None[Claim]()
	}
	parts := make([]string, len(roles))
	for i, r := range roles {
		parts[i] = r.value.value
	}

	return mo.Some(newStringClaim(scpClaim, NonEmptyString{value: strings.Join(parts, " ")}))
}

// RefreshTokenTypClaim returns a typ=Bearer claim.
func RefreshTokenTypClaim() Claim {
	return newStringClaim(typClaim, NonEmptyString{value: "Bearer"})
}

// Azpacr0Claim returns an azpacr=0 claim.
func Azpacr0Claim() Claim {
	return newStringClaim(azpacrClaim, NonEmptyString{value: "0"})
}

// Azpacls0Claim returns an azpacls=0 claim.
func Azpacls0Claim() Claim {
	return newStringClaim(azpaclsClaim, NonEmptyString{value: "0"})
}
