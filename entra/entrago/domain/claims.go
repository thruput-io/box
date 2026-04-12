package domain

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/mo"
)

// Claims is an immutable, ordered collection of typed claim key-payload pairs.
// The only way to add claims is via Append, which returns a new instance (copy-on-write).
type Claims struct {
	values map[claimKey]claimPayload
	order  []claimKey
}

// EmptyClaims returns a new empty Claims collection.
func EmptyClaims() Claims {
	return Claims{
		values: make(map[claimKey]claimPayload),
		order:  []claimKey{},
	}
}

// Append returns a new Claims with the given Claim merged in (copy-on-write, infallible).
//
// Merge rules (determined solely by claimKey.allowsArray):
//   - allowsArray=false: the incoming payload replaces any existing one (last wins).
//   - allowsArray=true:  payloads are merged into a single array, preserving insertion order.
func (c Claims) Append(cl Claim) Claims {
	newValues := cloneClaimMap(c.values)
	newOrder := cloneClaimOrder(c.order)

	ck := cl.key
	incoming := cl.payload

	existing, found := newValues[ck]
	if !found {
		newValues[ck] = incoming
		newOrder = append(newOrder, ck)

		return Claims{values: newValues, order: newOrder}
	}

	if !ck.allowsArray {
		newValues[ck] = incoming

		return Claims{values: newValues, order: newOrder}
	}

	// Collect all values from existing and incoming, then store as a single array payload.
	var all []claimValue
	switch existing.card {
	case claimCardinalityScalar:
		all = append(all, existing.one)
	case claimCardinalityArray:
		all = append(all, existing.many...)
	}
	switch incoming.card {
	case claimCardinalityScalar:
		all = append(all, incoming.one)
	case claimCardinalityArray:
		all = append(all, incoming.many...)
	}

	newValues[ck] = newArrayPayload(all)

	return Claims{values: newValues, order: newOrder}
}

// ContainsValue reports whether the claim for key k holds the given domain value.
func (c Claims) ContainsValue(k ValidClaimKey, v ClaimComparable) bool {
	payload, ok := c.values[k.key]
	if !ok {
		return false
	}

	return payload.contains(v.toClaimValue())
}

// ContainsStringValue reports whether the claim for key k holds the given raw string value.
// Returns false for an empty string or an absent key.
func (c Claims) ContainsStringValue(k ValidClaimKey, s string) bool {
	if s == "" {
		return false
	}
	payload, ok := c.values[k.key]
	if !ok {
		return false
	}

	return payload.contains(newClaimStringValue(NonEmptyString{value: s}))
}

// ContainsInt64Value reports whether the claim for key k holds the given raw int64 value.
func (c Claims) ContainsInt64Value(k ValidClaimKey, n int64) bool {
	payload, ok := c.values[k.key]
	if !ok {
		return false
	}

	return payload.contains(newClaimInt64Value(n))
}

// Subject returns the sub claim value, if present.
func (c Claims) Subject() mo.Option[Subject] {
	payload, ok := c.values[subClaim]
	if !ok || payload.card != claimCardinalityScalar {
		return mo.None[Subject]()
	}

	return mo.Some(Subject{value: payload.one.value})
}

// Nonce returns the nonce claim value, if present.
func (c Claims) Nonce() mo.Option[Nonce] {
	payload, ok := c.values[nonceClaim]
	if !ok || payload.card != claimCardinalityScalar {
		return mo.None[Nonce]()
	}

	return mo.Some(Nonce{value: payload.one.value})
}

// AuthorizedPartyClientID returns the azp claim value as a ClientID, if present and a valid UUID.
func (c Claims) AuthorizedPartyClientID() mo.Option[ClientID] {
	payload, ok := c.values[azpClaim]
	if !ok || payload.card != claimCardinalityScalar {
		return mo.None[ClientID]()
	}
	id := NewClientID(payload.one.value.value)
	if id.IsLeft() {
		return mo.None[ClientID]()
	}

	return mo.Some(id.MustRight())
}

// ScopeValues returns all scp claim values as a slice of ScopeValue.
// scp is stored as a space-separated scalar string; this method splits it into individual values.
// Returns an empty slice when the scp claim is absent.
func (c Claims) ScopeValues() []ScopeValue {
	payload, ok := c.values[scpClaim]
	if !ok || payload.card != claimCardinalityScalar {
		return []ScopeValue{}
	}
	parts := strings.Split(payload.one.value.value, " ")
	result := make([]ScopeValue, 0, len(parts))
	for _, part := range parts {
		if part == emptyString {
			continue
		}
		if sv, ok := NewScopeValue(part).Right(); ok {
			result = append(result, sv)
		}
	}

	return result
}

// From parses a jwt.MapClaims into a typed Claims collection.
// Unknown claim keys are silently skipped. Returns an error if any known claim's value
// cannot be converted to the expected type.
func From(raw jwt.MapClaims) mo.Either[Error, Claims] {
	out := EmptyClaims()

	for rawKey, rawValue := range raw {
		validKey, found := ValidClaim(rawKey).Get()
		if !found {
			continue
		}

		claimE := parseRawClaimValue(validKey, rawValue)

		claim, ok := claimE.Right()
		if !ok {
			return mo.Left[Error, Claims](claimE.MustLeft())
		}

		out = out.Append(claim)
	}

	return mo.Right[Error, Claims](out)
}

// FilterRefreshClaims returns a new Claims containing only the iss, aud, azp and tid claims.
// Returns None if any of the four required claims is absent.
func (c Claims) FilterRefreshClaims() mo.Option[Claims] {
	required := []claimKey{issClaim, audClaim, azpClaim, tidClaim}

	out := EmptyClaims()

	for _, requiredKey := range required {
		payload, found := c.values[requiredKey]
		if !found {
			return mo.None[Claims]()
		}

		out.values[requiredKey] = payload
		out.order = append(out.order, requiredKey)
	}

	return mo.Some(out)
}

// parseRawClaimValue converts a raw JWT any value into a typed Claim using the key's valueKind.
func parseRawClaimValue(key ValidClaimKey, raw any) mo.Either[Error, Claim] {
	switch typed := raw.(type) {
	case string:
		return key.NewStringClaim(typed)
	case []any:
		return parseRawStringArrayClaim(key, typed)
	default:
		nE := strictInt64FromAny(raw)
		if nE.IsLeft() {
			return mo.Left[Error, Claim](nE.MustLeft())
		}

		return key.NewInt64Claim(nE.MustRight())
	}
}

func parseRawStringArrayClaim(key ValidClaimKey, arr []any) mo.Either[Error, Claim] {
	elems := make([]string, 0, len(arr))

	for _, elem := range arr {
		s, ok := elem.(string)
		if !ok {
			return mo.Left[Error, Claim](ErrClaimTypeMismatch)
		}

		elems = append(elems, s)
	}

	return key.NewStringArrayClaim(elems)
}

// strictInt64FromAny converts a raw JWT numeric value (float64, int64, int, json.Number)
// into int64. Rejects fractional floats, NaN, Inf, and non-numeric types.
func strictInt64FromAny(value any) mo.Either[Error, int64] {
	switch typed := value.(type) {
	case int64:
		return mo.Right[Error, int64](typed)
	case int:
		return mo.Right[Error, int64](int64(typed))
	case float64:
		return strictInt64FromFloat64(typed)
	case json.Number:
		return strictInt64FromJSONNumber(typed)
	default:
		return mo.Left[Error, int64](ErrClaimTypeMismatch)
	}
}

func strictInt64FromFloat64(val float64) mo.Either[Error, int64] {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return mo.Left[Error, int64](ErrClaimValueInvalid)
	}

	if math.Trunc(val) != val {
		return mo.Left[Error, int64](ErrClaimValueInvalid)
	}

	if val < math.MinInt64 || val > math.MaxInt64 {
		return mo.Left[Error, int64](ErrClaimValueInvalid)
	}

	return mo.Right[Error, int64](int64(val))
}

func strictInt64FromJSONNumber(num json.Number) mo.Either[Error, int64] {
	parsed, err := num.Int64()
	if err != nil {
		return mo.Left[Error, int64](ErrClaimValueInvalid)
	}

	return mo.Right[Error, int64](parsed)
}

// ToJWT converts Claims to jwt.MapClaims for signing.
// This is the only place in the domain where claimKey strings are used as map keys.
func (c Claims) ToJWT() jwt.MapClaims {
	out := make(jwt.MapClaims, len(c.order))
	for _, k := range c.order {
		out[k.key] = payloadToJWTValue(c.values[k])
	}

	return out
}

func payloadToJWTValue(p claimPayload) any {
	switch p.card {
	case claimCardinalityScalar:
		return claimValueToJWTAny(p.one)
	case claimCardinalityArray:
		items := p.many
		result := make([]any, len(items))
		for i, v := range items {
			result[i] = claimValueToJWTAny(v)
		}

		return result
	default:
		panic("unreachable: unknown claimCardinality")
	}
}

func claimValueToJWTAny(v claimValue) any {
	switch v.kind {
	case claimValueKindString:
		return v.value.value
	case claimValueKindInt64:
		return v.n
	default:
		panic("unreachable: unknown claimValueKind")
	}
}
