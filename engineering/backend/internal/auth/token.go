package auth

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// jwksRefreshCooldown limits how often a cache-miss
	// can trigger a real HTTP fetch.
	jwksRefreshCooldown = 10 * time.Second
	// httpClientTimeout bounds outbound HTTP calls to ThunderID.
	httpClientTimeout = 10 * time.Second
	// jwtClockLeeway tolerates clock skew when validating exp/nbf.
	jwtClockLeeway = 30 * time.Second
	// jwksCacheMaxAge bounds how long a revoked key stays trusted.
	jwksCacheMaxAge = time.Hour
)

// Config holds ThunderID settings, loaded once from the environment in
// main.go and passed in explicitly — see NewVerifier.
type Config struct {
	BaseURL     string
	Issuer      string
	Audience    string
	InsecureTLS bool
}

// Validate reports which required fields are missing. InsecureTLS is
// intentionally excluded — unset means "use real TLS verification",
// which is a valid, safe default. A missing Issuer in particular must
// fail fast here, since jwt/v5 silently accepts a token with ANY issuer
// when WithIssuer is given an empty string.
func (c Config) Validate() error {
	var missing []string
	if c.BaseURL == "" {
		missing = append(missing, "THUNDER_BASE_URL")
	}
	if c.Issuer == "" {
		missing = append(missing, "THUNDER_ISSUER")
	}
	if c.Audience == "" {
		missing = append(missing, "THUNDER_AUDIENCE")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

// customClaims adds the space-delimited "scope" claim ThunderID issues
// (e.g. "openid book:create book:update") on top of the standard
// iss/sub/exp/aud claims, which jwt.RegisteredClaims already parses
// (including the string-or-array "aud" shape the JWT spec allows).
type customClaims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

type jwk struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// ErrJWKSUnavailable wraps a failure to fetch ThunderID's JWKS — distinct
// from an invalid token, so callers can respond 503 instead of 401.
var ErrJWKSUnavailable = errors.New("JWKS unavailable")

// Verifier verifies ThunderID-issued JWTs against a cached JWKS. Build one
// with NewVerifier at startup and share it across requests.
type Verifier struct {
	cfg        Config
	httpClient *http.Client

	cacheMu  sync.RWMutex // guards cache and cachedAt
	cache    []jwk
	cachedAt time.Time

	fetchMu     sync.Mutex // serializes JWKS fetches
	lastAttempt time.Time
}

// NewVerifier builds a Verifier from cfg, configuring the HTTP client's
// TLS verification per cfg.InsecureTLS.
func NewVerifier(cfg Config) *Verifier {
	client := &http.Client{Timeout: httpClientTimeout}
	if cfg.InsecureTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return &Verifier{cfg: cfg, httpClient: client}
}

func (v *Verifier) getJWKS() ([]jwk, error) {
	v.cacheMu.RLock()
	cache, cachedAt := v.cache, v.cachedAt
	v.cacheMu.RUnlock()
	if cache != nil && time.Since(cachedAt) < jwksCacheMaxAge {
		return cache, nil
	}
	return v.fetchJWKS()
}

// refreshJWKS re-fetches the JWKS, called when a token's kid isn't found
// in the cached set (ThunderID likely rotated its signing keys).
func (v *Verifier) refreshJWKS() ([]jwk, error) {
	return v.fetchJWKS()
}

// fetchJWKS does the actual HTTP fetch, skipping it if an attempt
// (successful or not) was made less than jwksRefreshCooldown ago.
func (v *Verifier) fetchJWKS() ([]jwk, error) {
	v.fetchMu.Lock()
	defer v.fetchMu.Unlock()

	if time.Since(v.lastAttempt) < jwksRefreshCooldown {
		v.cacheMu.RLock()
		cache := v.cache
		v.cacheMu.RUnlock()
		if cache != nil {
			return cache, nil
		}
		return nil, fmt.Errorf("%w: retry cooldown active", ErrJWKSUnavailable)
	}
	v.lastAttempt = time.Now()

	resp, err := v.httpClient.Get(v.cfg.BaseURL + "/oauth2/jwks")
	if err != nil {
		return nil, fmt.Errorf("%w: cannot reach ThunderID JWKS", ErrJWKSUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: ThunderID JWKS returned status %d", ErrJWKSUnavailable, resp.StatusCode)
	}

	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if json.NewDecoder(resp.Body).Decode(&doc) != nil || len(doc.Keys) == 0 {
		return nil, fmt.Errorf("%w: bad JWKS document", ErrJWKSUnavailable)
	}

	v.cacheMu.Lock()
	v.cache = doc.Keys
	v.cachedAt = time.Now()
	v.cacheMu.Unlock()
	return doc.Keys, nil
}

// Build an *rsa.PublicKey from a JWK's modulus (n) and exponent (e).
func translatePublicKeyFromJWK(k jwk) (*rsa.PublicKey, error) {
	nBytes, err1 := base64.RawURLEncoding.DecodeString(k.N)
	eBytes, err2 := base64.RawURLEncoding.DecodeString(k.E)
	if err1 != nil || err2 != nil || len(eBytes) > 8 {
		return nil, errors.New("bad JWK")
	}
	padded := make([]byte, 8)
	copy(padded[8-len(eBytes):], eBytes)
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(binary.BigEndian.Uint64(padded)),
	}, nil
}

// Full verification: signature + standard claims

// jwksKeyfunc resolves the RSA public key to verify a token's signature
// with, by looking up its "kid" header in ThunderID's JWKS. An unrecognized
// kid triggers one refresh before giving up, in case ThunderID rotated its
// signing keys since the cache was last populated.
func (v *Verifier) jwksKeyfunc(t *jwt.Token) (any, error) {
	kid, _ := t.Header["kid"].(string)

	keys, err := v.getJWKS()
	if err != nil {
		return nil, err
	}
	if key, ok := findJWK(keys, kid); ok {
		return translatePublicKeyFromJWK(key)
	}

	keys, err = v.refreshJWKS()
	if err != nil {
		return nil, err
	}
	if key, ok := findJWK(keys, kid); ok {
		return translatePublicKeyFromJWK(key)
	}
	return nil, errors.New("signing key not found")
}

func findJWK(keys []jwk, kid string) (jwk, bool) {
	for _, k := range keys {
		if k.Kid == kid {
			return k, true
		}
	}
	return jwk{}, false
}

func (v *Verifier) VerifyToken(raw string) (*customClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256"}), // algorithm pinning: the server decides, never the token.
		jwt.WithIssuer(v.cfg.Issuer),
		jwt.WithExpirationRequired(), // a token with no exp claim is rejected, not treated as never-expiring.
		jwt.WithLeeway(jwtClockLeeway),
		jwt.WithAudience(v.cfg.Audience),
	}

	var claims customClaims
	if _, err := jwt.ParseWithClaims(raw, &claims, v.jwksKeyfunc, opts...); err != nil {
		return nil, err
	}
	return &claims, nil
}
