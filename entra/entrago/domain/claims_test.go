package domain

import (
	"encoding/json"
	"testing"
)

// --- strictInt64FromAny ---

const (
	testTimestamp    = int64(1700000000)
	testTimestampF64 = float64(1700000000)
	testSmallInt     = int64(42)
	testRoleAdmin    = "Admin"
	testRoleReader   = "Reader"
	fmtGotWant       = "got %v, want %v"
	fmtCodeWant      = "code = %v, want %v"
)

func TestStrictInt64FromAny_Int64(t *testing.T) {
	t.Parallel()

	got := strictInt64FromAny(testTimestamp).MustRight()

	if got != testTimestamp {
		t.Fatalf(fmtGotWant, got, testTimestamp)
	}
}

func TestStrictInt64FromAny_Int(t *testing.T) {
	t.Parallel()

	got := strictInt64FromAny(int(testSmallInt)).MustRight()

	if got != testSmallInt {
		t.Fatalf(fmtGotWant, got, testSmallInt)
	}
}

func TestStrictInt64FromAny_Float64Integral(t *testing.T) {
	t.Parallel()

	got := strictInt64FromAny(testTimestampF64).MustRight()

	if got != testTimestamp {
		t.Fatalf(fmtGotWant, got, testTimestamp)
	}
}

func TestStrictInt64FromAny_Float64Fractional(t *testing.T) {
	t.Parallel()

	err := strictInt64FromAny(float64(1700000000.5)).MustLeft()

	if err.Code != ErrCodeClaimValueInvalid {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeClaimValueInvalid)
	}
}

func TestStrictInt64FromAny_JsonNumber(t *testing.T) {
	t.Parallel()

	got := strictInt64FromAny(json.Number("1700000000")).MustRight()

	if got != testTimestamp {
		t.Fatalf(fmtGotWant, got, testTimestamp)
	}
}

func TestStrictInt64FromAny_JsonNumberFractional(t *testing.T) {
	t.Parallel()

	err := strictInt64FromAny(json.Number("1700000000.5")).MustLeft()

	if err.Code != ErrCodeClaimValueInvalid {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeClaimValueInvalid)
	}
}

func TestStrictInt64FromAny_StringReturnsTypeMismatch(t *testing.T) {
	t.Parallel()

	err := strictInt64FromAny("not-a-number").MustLeft()

	if err.Code != ErrCodeClaimTypeMismatch {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeClaimTypeMismatch)
	}
}

// --- parseRawClaimValue ---

func TestParseRawClaimValue_StringKey(t *testing.T) {
	t.Parallel()

	key := ValidClaim("sub").MustGet()

	claim := parseRawClaimValue(key, "user-123").MustRight()

	if !claim.HasStringValue("user-123") {
		t.Fatal("expected claim to have value 'user-123'")
	}
}

func TestParseRawClaimValue_Int64Key(t *testing.T) {
	t.Parallel()

	key := ValidClaim("exp").MustGet()

	claim := parseRawClaimValue(key, testTimestampF64).MustRight()

	if !claim.HasInt64Value(testTimestamp) {
		t.Fatalf(fmtGotWant, claim, testTimestamp)
	}
}

func TestParseRawClaimValue_ArrayKey(t *testing.T) {
	t.Parallel()

	key := ValidClaim("roles").MustGet()

	claim := parseRawClaimValue(key, []any{testRoleAdmin, testRoleReader}).MustRight()

	if !claim.HasStringValue(testRoleAdmin) {
		t.Fatal("expected roles claim to contain 'Admin'")
	}

	if !claim.HasStringValue(testRoleReader) {
		t.Fatal("expected roles claim to contain 'Reader'")
	}
}

func TestParseRawClaimValue_EmptyStringReturnsErr(t *testing.T) {
	t.Parallel()

	key := ValidClaim("sub").MustGet()

	err := parseRawClaimValue(key, "").MustLeft()

	if err.Code != ErrCodeNonEmptyStringEmpty {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeNonEmptyStringEmpty)
	}
}

func TestParseRawClaimValue_FractionalFloat64ReturnsErr(t *testing.T) {
	t.Parallel()

	key := ValidClaim("exp").MustGet()

	err := parseRawClaimValue(key, float64(1700000000.5)).MustLeft()

	if err.Code != ErrCodeClaimValueInvalid {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeClaimValueInvalid)
	}
}

// --- Claims.From ---

func TestClaimsFrom_ScalarString(t *testing.T) {
	t.Parallel()

	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	raw := map[string]any{"tid": tenantID.Value()}

	claims := From(raw).MustRight()

	key := ValidClaim("tid").MustGet()

	if !claims.ContainsValue(key, tenantID) {
		t.Fatal("expected claims to contain tenantID")
	}
}

func TestClaimsFrom_ScalarInt64(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"exp": testTimestampF64}

	claims := From(raw).MustRight()

	key := ValidClaim("exp").MustGet()

	if !claims.ContainsInt64Value(key, testTimestamp) {
		t.Fatal("expected claims to contain exp")
	}
}

func TestClaimsFrom_StringArray(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"roles": []any{testRoleAdmin, testRoleReader}}

	claims := From(raw).MustRight()

	key := ValidClaim("roles").MustGet()

	if !claims.ContainsStringValue(key, testRoleAdmin) {
		t.Fatal("expected claims to contain Admin role")
	}
}

func TestClaimsFrom_UnknownKeySkipped(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"unknown_xyz": "value", "tid": "11111111-1111-1111-1111-111111111111"}

	claims := From(raw).MustRight()

	key := ValidClaim("tid").MustGet()

	if !claims.ContainsStringValue(key, "11111111-1111-1111-1111-111111111111") {
		t.Fatal("expected tid to be present")
	}
}

func TestClaimsFrom_InvalidValueReturnsErr(t *testing.T) {
	t.Parallel()

	raw := map[string]any{"sub": ""}

	err := From(raw).MustLeft()

	if err.Code != ErrCodeNonEmptyStringEmpty {
		t.Fatalf(fmtCodeWant, err.Code, ErrCodeNonEmptyStringEmpty)
	}
}

// --- Claims.FilterRefreshClaims ---

func TestFilterRefreshClaims_AllPresent(t *testing.T) {
	t.Parallel()

	issuer := NewIssuer("https://issuer.example.com").MustRight()
	clientID := NewClientID("22222222-2222-2222-2222-222222222222").MustRight()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()

	claims := EmptyClaims().
		Append(issuer.AsClaim()).
		Append(issuer.AsAudClaim()).
		Append(clientID.AsClaim()).
		Append(tenantID.AsClaim())

	filtered, ok := claims.FilterRefreshClaims().Get()

	if !ok {
		t.Fatal("expected FilterRefreshClaims to return Some")
	}

	issKey := ValidClaim("iss").MustGet()
	audKey := ValidClaim("aud").MustGet()
	azpKey := ValidClaim("azp").MustGet()
	tidKey := ValidClaim("tid").MustGet()

	if !filtered.ContainsValue(issKey, issuer) {
		t.Fatal("expected filtered claims to contain iss")
	}

	if !filtered.ContainsValue(audKey, issuer) {
		t.Fatal("expected filtered claims to contain aud")
	}

	if !filtered.ContainsValue(azpKey, clientID) {
		t.Fatal("expected filtered claims to contain azp")
	}

	if !filtered.ContainsValue(tidKey, tenantID) {
		t.Fatal("expected filtered claims to contain tid")
	}
}

func TestFilterRefreshClaims_MissingClaimReturnsNone(t *testing.T) {
	t.Parallel()

	issuer := NewIssuer("https://issuer.example.com").MustRight()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()

	// azp and aud intentionally absent
	claims := EmptyClaims().
		Append(issuer.AsClaim()).
		Append(tenantID.AsClaim())

	if claims.FilterRefreshClaims().IsPresent() {
		t.Fatal("expected FilterRefreshClaims to return None when azp/aud absent")
	}
}

func TestFilterRefreshClaims_ExcludesOtherClaims(t *testing.T) {
	t.Parallel()

	issuer := NewIssuer("https://issuer.example.com").MustRight()
	clientID := NewClientID("22222222-2222-2222-2222-222222222222").MustRight()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	nonce := NewNonce("test-nonce").MustRight()

	claims := EmptyClaims().
		Append(issuer.AsClaim()).
		Append(issuer.AsAudClaim()).
		Append(clientID.AsClaim()).
		Append(tenantID.AsClaim()).
		Append(nonce.AsClaim())

	filtered := claims.FilterRefreshClaims().MustGet()

	nonceKey := ValidClaim("nonce").MustGet()

	if filtered.ContainsValue(nonceKey, nonce) {
		t.Fatal("expected filtered claims to exclude nonce")
	}
}

// --- Claim.HasValue (domain types via ClaimComparable) ---

func TestClaimHasValue_TenantID(t *testing.T) {
	t.Parallel()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	claim := tenantID.AsClaim()
	if !claim.HasValue(tenantID) {
		t.Fatalf("expected claim to have tenantID value")
	}
}

func TestClaimHasValue_ClientID(t *testing.T) {
	t.Parallel()
	clientID := NewClientID("22222222-2222-2222-2222-222222222222").MustRight()
	claim := clientID.AsClaim()
	if !claim.HasValue(clientID) {
		t.Fatalf("expected claim to have clientID value")
	}
}

func TestClaimHasValue_Nonce(t *testing.T) {
	t.Parallel()
	nonce := NewNonce("abc-nonce").MustRight()
	claim := nonce.AsClaim()
	if !claim.HasValue(nonce) {
		t.Fatalf("expected claim to have nonce value")
	}
}

func TestClaimHasValue_Issuer(t *testing.T) {
	t.Parallel()
	issuer := NewIssuer("https://issuer.example.com").MustRight()
	claim := issuer.AsClaim()
	if !claim.HasValue(issuer) {
		t.Fatalf("expected claim to have issuer value")
	}
}

func TestClaimHasValue_Subject(t *testing.T) {
	t.Parallel()
	subject := NewUserID("33333333-3333-3333-3333-333333333333").MustRight().AsSubject()
	claim := subject.AsClaim()
	if !claim.HasValue(subject) {
		t.Fatalf("expected claim to have subject value")
	}
}

func TestClaimHasValue_TokenVersion(t *testing.T) {
	t.Parallel()
	version := NewTokenVersion("2.0").MustRight()
	claim := version.AsClaim()
	if !claim.HasValue(version) {
		t.Fatalf("expected claim to have version value")
	}
}

func TestClaimHasValue_DisplayName(t *testing.T) {
	t.Parallel()
	name := NewDisplayName("Mock User").MustRight()
	claim := name.AsClaim()
	if !claim.HasValue(name) {
		t.Fatalf("expected claim to have displayName value")
	}
}

func TestClaimHasValue_Email(t *testing.T) {
	t.Parallel()
	email := NewEmail("user@example.com").MustRight()
	claim := email.AsClaim()
	if !claim.HasValue(email) {
		t.Fatalf("expected claim to have email value")
	}
}

func TestClaimHasValue_ReturnsFalseForWrongValue(t *testing.T) {
	t.Parallel()
	nonce := NewNonce("correct").MustRight()
	other := NewNonce("wrong").MustRight()
	claim := nonce.AsClaim()
	if claim.HasValue(other) {
		t.Fatalf("expected HasValue to return false for wrong value")
	}
}

// --- Claim.HasStringValue / HasInt64Value (raw primitives) ---

func TestClaimHasStringValue(t *testing.T) {
	t.Parallel()
	nonce := NewNonce("raw-string").MustRight()
	claim := nonce.AsClaim()
	if !claim.HasStringValue("raw-string") {
		t.Fatalf("expected HasStringValue to return true")
	}
}

func TestClaimHasStringValue_EmptyReturnsFalse(t *testing.T) {
	t.Parallel()
	nonce := NewNonce("something").MustRight()
	claim := nonce.AsClaim()
	if claim.HasStringValue("") {
		t.Fatalf("expected HasStringValue(\"\") to return false")
	}
}

func TestClaimHasInt64Value(t *testing.T) {
	t.Parallel()
	claim := NewJwtNumericDateSeconds(1700000000).AsExpClaim()
	if !claim.HasInt64Value(1700000000) {
		t.Fatalf("expected HasInt64Value to return true")
	}
}

func TestClaimHasInt64Value_ReturnsFalseForWrongValue(t *testing.T) {
	t.Parallel()
	claim := NewJwtNumericDateSeconds(1700000000).AsExpClaim()
	if claim.HasInt64Value(9999999999) {
		t.Fatalf("expected HasInt64Value to return false for wrong value")
	}
}

// --- AsClaim() variants ---

func TestAsOidClaim(t *testing.T) {
	t.Parallel()
	subject := NewUserID("33333333-3333-3333-3333-333333333333").MustRight().AsSubject()
	claim := subject.AsOidClaim()
	if !claim.HasValue(subject) {
		t.Fatalf("expected oid claim to have subject value")
	}
}

func TestAsAppidClaim(t *testing.T) {
	t.Parallel()
	clientID := NewClientID("22222222-2222-2222-2222-222222222222").MustRight()
	claim := clientID.AsAppidClaim()
	if !claim.HasValue(clientID) {
		t.Fatalf("expected appid claim to have clientID value")
	}
}

func TestClientIDAsAudClaim(t *testing.T) {
	t.Parallel()
	clientID := NewClientID("22222222-2222-2222-2222-222222222222").MustRight()
	claim := clientID.AsAudClaim()
	if !claim.HasValue(clientID) {
		t.Fatalf("expected aud claim to have clientID value")
	}
}

func TestIssuerAsAudClaim(t *testing.T) {
	t.Parallel()
	issuer := NewIssuer("https://aud.example.com").MustRight()
	claim := issuer.AsAudClaim()
	if !claim.HasValue(issuer) {
		t.Fatalf("expected aud claim to have issuer value")
	}
}

// --- Named array constructors ---

func TestRolesClaimFrom(t *testing.T) {
	t.Parallel()
	role := NewRoleValue("Reader").MustRight()
	arr := NewNonEmptyArray(role).MustRight()
	claim := RolesClaimFrom(arr)
	if !claim.HasValue(role) {
		t.Fatalf("expected roles claim to contain role value")
	}
}

func TestGroupsClaimFrom(t *testing.T) {
	t.Parallel()
	groupID := NewGroupID("44444444-4444-4444-4444-444444444444").MustRight()
	arr := NewNonEmptyArray(groupID).MustRight()
	claim := GroupsClaimFrom(arr)
	if !claim.HasStringValue(groupID.Value()) {
		t.Fatalf("expected groups claim to contain group UUID string")
	}
}

func TestScpClaimFrom(t *testing.T) {
	t.Parallel()
	scope := NewScopeValue("read").MustRight()
	arr := NewNonEmptyArray(scope).MustRight()
	claim := ScpClaimFrom(arr)
	if !claim.HasValue(scope) {
		t.Fatalf("expected scp claim to contain scope value")
	}
}

// --- Constant claims ---

func TestRefreshTokenTypClaim(t *testing.T) {
	t.Parallel()
	claim := RefreshTokenTypClaim()
	if !claim.HasStringValue("Bearer") {
		t.Fatalf("expected typ claim to have value 'Bearer'")
	}
}

func TestAzpacr0Claim(t *testing.T) {
	t.Parallel()
	claim := Azpacr0Claim()
	if !claim.HasStringValue("0") {
		t.Fatalf("expected azpacr claim to have value '0'")
	}
}

// --- ValidClaim + typed ingress constructors ---

func TestValidClaim_KnownKeyReturnsKey(t *testing.T) {
	t.Parallel()
	opt := ValidClaim("sub")
	if opt.IsAbsent() {
		t.Fatal("expected ValidClaim('sub') to return a key")
	}
}

func TestValidClaim_UnknownKeyReturnsNone(t *testing.T) {
	t.Parallel()
	opt := ValidClaim("unknown_claim_xyz")
	if opt.IsPresent() {
		t.Fatal("expected ValidClaim('unknown_claim_xyz') to return None")
	}
}

func TestNewStringClaim_Success(t *testing.T) {
	t.Parallel()
	key := ValidClaim("sub").MustGet()
	claim := key.NewStringClaim("user123").MustRight()
	if !claim.HasStringValue("user123") {
		t.Fatal("expected claim to have value 'user123'")
	}
}

func TestNewStringClaim_EmptyStringReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("sub").MustGet()
	err := key.NewStringClaim("").MustLeft()
	if err.Code != ErrCodeNonEmptyStringEmpty {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeNonEmptyStringEmpty)
	}
}

func TestNewStringClaim_WrongKindReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("exp").MustGet() // int64 key
	err := key.NewStringClaim("not-a-number").MustLeft()
	if err.Code != ErrCodeClaimTypeMismatch {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeClaimTypeMismatch)
	}
}

func TestNewInt64Claim_Success(t *testing.T) {
	t.Parallel()
	key := ValidClaim("exp").MustGet()
	claim := key.NewInt64Claim(1700000000).MustRight()
	if !claim.HasInt64Value(1700000000) {
		t.Fatal("expected exp claim to have value 1700000000")
	}
}

func TestNewInt64Claim_WrongKindReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("sub").MustGet() // string key
	err := key.NewInt64Claim(42).MustLeft()
	if err.Code != ErrCodeClaimTypeMismatch {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeClaimTypeMismatch)
	}
}

func TestNewStringArrayClaim_Success(t *testing.T) {
	t.Parallel()
	key := ValidClaim("roles").MustGet()
	claim := key.NewStringArrayClaim([]string{"Admin", "Reader"}).MustRight()
	if !claim.HasStringValue("Admin") {
		t.Fatal("expected roles claim to contain 'Admin'")
	}
	if !claim.HasStringValue("Reader") {
		t.Fatal("expected roles claim to contain 'Reader'")
	}
}

func TestNewStringArrayClaim_EmptySliceReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("roles").MustGet()
	err := key.NewStringArrayClaim([]string{}).MustLeft()
	if err.Code != ErrCodeNonEmptyArrayEmpty {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeNonEmptyArrayEmpty)
	}
}

func TestNewStringArrayClaim_ScalarKeyReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("sub").MustGet() // allowsArray: false
	err := key.NewStringArrayClaim([]string{"value"}).MustLeft()
	if err.Code != ErrCodeClaimShapeInvalid {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeClaimShapeInvalid)
	}
}

func TestNewStringArrayClaim_EmptyElementReturnsErr(t *testing.T) {
	t.Parallel()
	key := ValidClaim("roles").MustGet()
	err := key.NewStringArrayClaim([]string{"Admin", ""}).MustLeft()
	if err.Code != ErrCodeNonEmptyStringEmpty {
		t.Fatalf("code = %v, want %v", err.Code, ErrCodeNonEmptyStringEmpty)
	}
}

// --- Claims.Append + ContainsValue ---

func TestClaimsAppend_ScalarInsert(t *testing.T) {
	t.Parallel()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	key := ValidClaim("tid").MustGet()

	claims := EmptyClaims().Append(tenantID.AsClaim())

	if !claims.ContainsValue(key, tenantID) {
		t.Fatalf("expected claims to contain tenantID")
	}
}

func TestClaimsAppend_ScalarReplaces(t *testing.T) {
	t.Parallel()
	nonce1 := NewNonce("first").MustRight()
	nonce2 := NewNonce("second").MustRight()
	key := ValidClaim("nonce").MustGet()

	claims := EmptyClaims().
		Append(nonce1.AsClaim()).
		Append(nonce2.AsClaim())

	if claims.ContainsValue(key, nonce1) {
		t.Fatalf("expected first nonce to be replaced")
	}
	if !claims.ContainsValue(key, nonce2) {
		t.Fatalf("expected second nonce to be present")
	}
}

func TestClaimsAppend_ArrayMerges(t *testing.T) {
	t.Parallel()
	role1 := NewRoleValue("Reader").MustRight()
	role2 := NewRoleValue("Writer").MustRight()
	key := ValidClaim("roles").MustGet()

	claims := EmptyClaims().
		Append(RolesClaimFrom(NewNonEmptyArray(role1).MustRight())).
		Append(RolesClaimFrom(NewNonEmptyArray(role2).MustRight()))

	if !claims.ContainsValue(key, role1) {
		t.Fatalf("expected merged claims to contain role1")
	}
	if !claims.ContainsValue(key, role2) {
		t.Fatalf("expected merged claims to contain role2")
	}
}

func TestClaimsAppend_ImmutableOriginal(t *testing.T) {
	t.Parallel()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	key := ValidClaim("tid").MustGet()

	original := EmptyClaims()
	updated := original.Append(tenantID.AsClaim())

	if original.ContainsValue(key, tenantID) {
		t.Fatalf("Append must not modify the original Claims")
	}
	if !updated.ContainsValue(key, tenantID) {
		t.Fatalf("expected updated claims to contain tenantID")
	}
}

func TestClaimsContainsStringValue(t *testing.T) {
	t.Parallel()
	key := ValidClaim("nonce").MustGet()
	nonce := NewNonce("my-nonce").MustRight()
	claims := EmptyClaims().Append(nonce.AsClaim())

	if !claims.ContainsStringValue(key, "my-nonce") {
		t.Fatalf("expected ContainsStringValue to return true")
	}
	if claims.ContainsStringValue(key, "wrong") {
		t.Fatalf("expected ContainsStringValue to return false for wrong value")
	}
}

func TestClaimsContainsInt64Value(t *testing.T) {
	t.Parallel()
	key := ValidClaim("exp").MustGet()
	claims := EmptyClaims().Append(NewJwtNumericDateSeconds(1700000000).AsExpClaim())

	if !claims.ContainsInt64Value(key, 1700000000) {
		t.Fatalf("expected ContainsInt64Value to return true")
	}
	if claims.ContainsInt64Value(key, 9999999999) {
		t.Fatalf("expected ContainsInt64Value to return false for wrong value")
	}
}

// --- Claims.ToJWT round-trip ---

func TestClaimsToJWT_ScalarString(t *testing.T) {
	t.Parallel()
	tenantID := NewTenantID("11111111-1111-1111-1111-111111111111").MustRight()
	claims := EmptyClaims().Append(tenantID.AsClaim())
	jwt := claims.ToJWT()

	v, ok := jwt["tid"]
	if !ok {
		t.Fatalf("expected 'tid' in JWT map")
	}
	if v != tenantID.Value() {
		t.Fatalf("tid = %v, want %v", v, tenantID.Value())
	}
}

func TestClaimsToJWT_ScalarInt64(t *testing.T) {
	t.Parallel()
	claims := EmptyClaims().Append(NewJwtNumericDateSeconds(1700000000).AsExpClaim())
	jwt := claims.ToJWT()

	v, ok := jwt["exp"]
	if !ok {
		t.Fatalf("expected 'exp' in JWT map")
	}
	if v != int64(1700000000) {
		t.Fatalf("exp = %v, want %v", v, int64(1700000000))
	}
}

func TestClaimsToJWT_StringArray(t *testing.T) {
	t.Parallel()
	role := NewRoleValue("Admin").MustRight()
	claims := EmptyClaims().Append(RolesClaimFrom(NewNonEmptyArray(role).MustRight()))
	jwt := claims.ToJWT()

	v, ok := jwt["roles"]
	if !ok {
		t.Fatalf("expected 'roles' in JWT map")
	}
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected roles to be []any, got %T", v)
	}
	if len(arr) != 1 || arr[0] != "Admin" {
		t.Fatalf("roles = %v, want [Admin]", arr)
	}
}

// --- Claims read accessors ---

func TestClaimsSubject_Present(t *testing.T) {
	t.Parallel()
	subject := NewUserID("33333333-3333-3333-3333-333333333333").MustRight().AsSubject()
	claims := EmptyClaims().Append(subject.AsClaim())

	got := claims.Subject()
	if got.IsAbsent() {
		t.Fatalf("expected Subject() to return a value")
	}
	if got.MustGet().Value() != subject.Value() {
		t.Fatalf("subject = %v, want %v", got.MustGet().Value(), subject.Value())
	}
}

func TestClaimsSubject_Absent(t *testing.T) {
	t.Parallel()
	claims := EmptyClaims()
	if claims.Subject().IsPresent() {
		t.Fatalf("expected Subject() to be absent for empty Claims")
	}
}

func TestClaimsNonce_Present(t *testing.T) {
	t.Parallel()
	nonce := NewNonce("test-nonce").MustRight()
	claims := EmptyClaims().Append(nonce.AsClaim())

	got := claims.Nonce()
	if got.IsAbsent() {
		t.Fatalf("expected Nonce() to return a value")
	}
	if got.MustGet().Value() != nonce.Value() {
		t.Fatalf("nonce = %v, want %v", got.MustGet().Value(), nonce.Value())
	}
}
