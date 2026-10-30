package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// test helpers shared with middleware_test.go

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test RSA key: %v", err)
	}
	return key
}

func jwkFromPublicKey(kid string, pub *rsa.PublicKey) jwk {
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	return jwk{
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}
}

func newJWKSServer(t *testing.T, keys []jwk) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func signTestToken(t *testing.T, key *rsa.PrivateKey, kid string, claims customClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("signing test token: %v", err)
	}
	return signed
}

// newTestVerifier builds a Verifier with its own private config and JWKS
// cache — no shared package state, so no cleanup is needed between tests.
func newTestVerifier(baseURL, issuer, audience string) *Verifier {
	return NewVerifier(Config{BaseURL: baseURL, Issuer: issuer, Audience: audience})
}

// Config.Validate

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL, issuer, audience string
		wantErrContains           string
	}{
		{"all set", "http://thunder", "http://thunder", "aud", ""},
		{"missing base url", "", "http://thunder", "aud", "THUNDER_BASE_URL"},
		{"missing issuer", "http://thunder", "", "aud", "THUNDER_ISSUER"},
		{"missing audience", "http://thunder", "http://thunder", "", "THUNDER_AUDIENCE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{BaseURL: tt.baseURL, Issuer: tt.issuer, Audience: tt.audience}

			err := cfg.Validate()
			if tt.wantErrContains == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
			}
		})
	}
}

// translatePublicKeyFromJWK

func TestTranslatePublicKeyFromJWK_RoundTrips(t *testing.T) {
	key := generateTestKey(t)
	j := jwkFromPublicKey("kid1", &key.PublicKey)

	pub, err := translatePublicKeyFromJWK(j)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pub.E != key.PublicKey.E || pub.N.Cmp(key.PublicKey.N) != 0 {
		t.Errorf("translated key does not match the original")
	}
}

func TestTranslatePublicKeyFromJWK_BadBase64(t *testing.T) {
	_, err := translatePublicKeyFromJWK(jwk{N: "not valid base64!!", E: "AQAB"})
	if err == nil {
		t.Fatal("expected an error for an undecodable N")
	}
}

func TestTranslatePublicKeyFromJWK_OversizedExponent(t *testing.T) {
	eBytes := make([]byte, 9) // more than the 8 bytes this function supports
	for i := range eBytes {
		eBytes[i] = 1
	}
	j := jwk{
		N: base64.RawURLEncoding.EncodeToString([]byte{1, 2, 3}),
		E: base64.RawURLEncoding.EncodeToString(eBytes),
	}

	_, err := translatePublicKeyFromJWK(j)
	if err == nil {
		t.Fatal("expected an error for an oversized exponent")
	}
}

// findJWK 

func TestFindJWK(t *testing.T) {
	keys := []jwk{{Kid: "a"}, {Kid: "b"}}

	if k, ok := findJWK(keys, "b"); !ok || k.Kid != "b" {
		t.Errorf("expected to find kid %q", "b")
	}
	if _, ok := findJWK(keys, "missing"); ok {
		t.Error("expected not to find an unknown kid")
	}
}

// fetchJWKS / getJWKS / refreshJWKS

func TestGetJWKS_Success(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "issuer", "aud")

	keys, err := v.getJWKS()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 1 || keys[0].Kid != "kid1" {
		t.Errorf("unexpected keys: %+v", keys)
	}
}

func TestGetJWKS_FreshCache_DoesNotRefetch(t *testing.T) {
	var hits int
	key := generateTestKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []jwk{jwkFromPublicKey("kid1", &key.PublicKey)}})
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")

	if _, err := v.getJWKS(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := v.getJWKS(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits != 1 {
		t.Errorf("expected the warm, fresh cache to be reused, got %d hits", hits)
	}
}

func TestGetJWKS_StaleCache_ForcesRefresh(t *testing.T) {
	var hits int
	key := generateTestKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []jwk{jwkFromPublicKey("kid1", &key.PublicKey)}})
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")

	if _, err := v.getJWKS(); err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected 1 hit after first call, got %d", hits)
	}

	// Simulate the cache having aged past jwksCacheMaxAge, and the
	// refresh cooldown no longer blocking a new attempt.
	v.cacheMu.Lock()
	v.cachedAt = time.Now().Add(-jwksCacheMaxAge - time.Second)
	v.cacheMu.Unlock()
	v.lastAttempt = time.Now().Add(-jwksRefreshCooldown - time.Second)

	if _, err := v.getJWKS(); err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if hits != 2 {
		t.Errorf("expected a stale cache to force a second fetch, got %d hits", hits)
	}
}

func TestFetchJWKSLocked_NonOKStatus_DoesNotUpdateCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")
	v.cache = []jwk{{Kid: "existing"}}

	_, err := v.fetchJWKS()
	if err == nil {
		t.Fatal("expected an error for a non-200 response")
	}

	if len(v.cache) != 1 || v.cache[0].Kid != "existing" {
		t.Errorf("cache should be untouched after a failed fetch, got %+v", v.cache)
	}
}

func TestFetchJWKSLocked_EmptyKeys_Rejected(t *testing.T) {
	srv := newJWKSServer(t, []jwk{})
	v := newTestVerifier(srv.URL, "issuer", "aud")
	v.cache = []jwk{{Kid: "existing"}}

	_, err := v.fetchJWKS()
	if err == nil {
		t.Fatal("expected an error for an empty keys list")
	}

	if len(v.cache) != 1 || v.cache[0].Kid != "existing" {
		t.Errorf("cache should be untouched after a rejected fetch, got %+v", v.cache)
	}
}

func TestFetchJWKSLocked_FailureRespectsCooldown(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")

	if _, err := v.fetchJWKS(); err == nil {
		t.Fatal("expected an error on first attempt")
	}
	if _, err := v.fetchJWKS(); err == nil {
		t.Fatal("expected an error on second attempt (still within cooldown)")
	}
	if hits != 1 {
		t.Errorf("expected the cooldown to prevent a second fetch after a failure, got %d hits", hits)
	}
}

func TestGetJWKS_EmptyCache_RespectsCooldownAfterFailure(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")

	if _, err := v.getJWKS(); err == nil {
		t.Fatal("expected an error on first call (IdP down)")
	}
	if _, err := v.getJWKS(); err == nil {
		t.Fatal("expected an error on second call (still within cooldown)")
	}
	if hits != 1 {
		t.Errorf("expected the cooldown to prevent a second fetch from getJWKS on an empty cache, got %d hits", hits)
	}
}

func TestRefreshJWKS_RespectsCooldown(t *testing.T) {
	var hits int
	key := generateTestKey(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []jwk{jwkFromPublicKey("kid1", &key.PublicKey)}})
	}))
	t.Cleanup(srv.Close)
	v := newTestVerifier(srv.URL, "issuer", "aud")

	if _, err := v.refreshJWKS(); err != nil {
		t.Fatalf("unexpected error on first refresh: %v", err)
	}
	if _, err := v.refreshJWKS(); err != nil {
		t.Fatalf("unexpected error on second refresh: %v", err)
	}
	if hits != 1 {
		t.Errorf("expected the cooldown to prevent a second fetch, got %d hits", hits)
	}
}

// verifyToken (full signature + claims verification)

func validTestClaims(issuer, audience, subject, scope string) customClaims {
	now := time.Now()
	return customClaims{
		Scope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	raw := signTestToken(t, key, "kid1", validTestClaims("test-issuer", "test-audience", "user-1", "book:create book:update"))

	claims, err := v.VerifyToken(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Errorf("expected subject %q, got %q", "user-1", claims.Subject)
	}
	if claims.Scope != "book:create book:update" {
		t.Errorf("unexpected scope: %q", claims.Scope)
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	claims := validTestClaims("test-issuer", "test-audience", "user-1", "")
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	raw := signTestToken(t, key, "kid1", claims)

	if _, err := v.VerifyToken(raw); err == nil {
		t.Fatal("expected an error for an expired token")
	}
}

func TestVerifyToken_WrongIssuer(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	raw := signTestToken(t, key, "kid1", validTestClaims("someone-else", "test-audience", "user-1", ""))

	if _, err := v.VerifyToken(raw); err == nil {
		t.Fatal("expected an error for a mismatched issuer")
	}
}

func TestVerifyToken_WrongAudience(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	raw := signTestToken(t, key, "kid1", validTestClaims("test-issuer", "someone-else", "user-1", ""))

	if _, err := v.VerifyToken(raw); err == nil {
		t.Fatal("expected an error for a mismatched audience")
	}
}

func TestVerifyToken_UnknownKid(t *testing.T) {
	key := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &key.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	raw := signTestToken(t, key, "kid-does-not-exist", validTestClaims("test-issuer", "test-audience", "user-1", ""))

	if _, err := v.VerifyToken(raw); err == nil {
		t.Fatal("expected an error for an unrecognized kid")
	}
}

func TestVerifyToken_BadSignature(t *testing.T) {
	realKey := generateTestKey(t)
	wrongKey := generateTestKey(t)
	srv := newJWKSServer(t, []jwk{jwkFromPublicKey("kid1", &realKey.PublicKey)})
	v := newTestVerifier(srv.URL, "test-issuer", "test-audience")

	// Signed with wrongKey but claims kid1, whose published JWK is realKey's
	// public key — the signature must not verify.
	raw := signTestToken(t, wrongKey, "kid1", validTestClaims("test-issuer", "test-audience", "user-1", ""))

	if _, err := v.VerifyToken(raw); err == nil {
		t.Fatal("expected an error for a signature that doesn't match the published key")
	}
}
