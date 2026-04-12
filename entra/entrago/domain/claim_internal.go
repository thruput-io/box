package domain

import (
	"maps"
)

type claimValueKind uint8

const (
	claimValueKindString claimValueKind = iota + 1
	claimValueKindInt64
)

// claimValue is a closed tagged union holding either a string or an int64 claim value.
// Exactly one of value (string) or n (int64) is active, determined by kind.
type claimValue struct {
	kind  claimValueKind
	value NonEmptyString // active when kind == claimValueKindString
	n     int64          // active when kind == claimValueKindInt64
}

func newClaimStringValue(v NonEmptyString) claimValue {
	return claimValue{kind: claimValueKindString, value: v, n: 0}
}

func newClaimInt64Value(n int64) claimValue {
	return claimValue{kind: claimValueKindInt64, value: NonEmptyString{value: ""}, n: n}
}

func (v claimValue) equals(other claimValue) bool {
	if v.kind != other.kind {
		return false
	}

	switch v.kind {
	case claimValueKindString:
		return v.value.value == other.value.value
	case claimValueKindInt64:
		return v.n == other.n
	default:
		return false
	}
}

type claimCardinality uint8

const (
	claimCardinalityScalar claimCardinality = iota + 1
	claimCardinalityArray
)

// claimPayload holds either a scalar or array of claimValues.
// card determines which variant is active.
// many is a plain slice so that zero-element array payloads (e.g. roles=[]) are representable for JWT egress.
type claimPayload struct {
	card claimCardinality
	one  claimValue
	many []claimValue
}

func newScalarPayload(v claimValue) claimPayload {
	return claimPayload{card: claimCardinalityScalar, one: v, many: nil}
}

func newArrayPayload(values []claimValue) claimPayload {
	return claimPayload{
		card: claimCardinalityArray,
		one:  claimValue{kind: claimValueKindString, value: NonEmptyString{value: ""}, n: 0},
		many: values,
	}
}

func (p claimPayload) contains(v claimValue) bool {
	switch p.card {
	case claimCardinalityScalar:
		return p.one.equals(v)
	case claimCardinalityArray:
		for _, item := range p.many {
			if item.equals(v) {
				return true
			}
		}

		return false
	default:
		panic("unreachable: unknown claimCardinality")
	}
}

func cloneClaimMap(src map[claimKey]claimPayload) map[claimKey]claimPayload {
	if src == nil {
		return nil
	}

	dst := make(map[claimKey]claimPayload, len(src))
	maps.Copy(dst, src)

	return dst
}

func cloneClaimOrder(src []claimKey) []claimKey {
	if src == nil {
		return nil
	}

	dst := make([]claimKey, len(src))
	copy(dst, src)

	return dst
}
