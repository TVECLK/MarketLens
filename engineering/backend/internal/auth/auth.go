package auth

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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

var (
	jwksMu          sync.Mutex
	jwksCache       []jwk
	jwksLastRefresh time.Time
)

// jwksRefreshCooldown limits how often a cache-miss (an unrecognized kid)
// can trigger a real HTTP fetch, so a token with a bogus kid can't be used
// to hammer ThunderID's JWKS endpoint on every request.
const jwksRefreshCooldown = 10 * time.Second

func newHTTPClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}

	if os.Getenv("THUNDER_INSECURE_TLS") == "true" {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return client
}

func getJWKS() ([]jwk, error) {
	jwksMu.Lock()
	defer jwksMu.Unlock()
	if jwksCache != nil {
		return jwksCache, nil
	}
	return fetchJWKSLocked()
}

// refreshJWKS re-fetches the JWKS unconditionally, replacing the cache —
// called when a token's kid isn't found in the cached set, which usually
// means ThunderID rotated its signing keys since the last fetch.
func refreshJWKS() ([]jwk, error) {
	jwksMu.Lock()
	defer jwksMu.Unlock()
	if time.Since(jwksLastRefresh) < jwksRefreshCooldown {
		return jwksCache, nil
	}
	return fetchJWKSLocked()
}

// fetchJWKSLocked does the actual HTTP fetch and cache update. Callers must
// hold jwksMu.
func fetchJWKSLocked() ([]jwk, error) {
	client := newHTTPClient()

	resp, err := client.Get(os.Getenv("THUNDER_BASE_URL") + "/oauth2/jwks")
	if err != nil {
		return nil, errors.New("cannot reach ThunderID JWKS")
	}
	defer resp.Body.Close()

	var doc struct {
		Keys []jwk `json:"keys"`
	}
	if json.NewDecoder(resp.Body).Decode(&doc) != nil {
		return nil, errors.New("bad JWKS document")
	}
	jwksCache = doc.Keys
	jwksLastRefresh = time.Now()
	return jwksCache, nil
}

// Build an *rsa.PublicKey from a JWK's modulus (n) and exponent (e).
func publicKeyFromJWK(k jwk) (*rsa.PublicKey, error) {
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
func jwksKeyfunc(t *jwt.Token) (any, error) {
	kid, _ := t.Header["kid"].(string)

	keys, err := getJWKS()
	if err != nil {
		return nil, err
	}
	if key, ok := findJWK(keys, kid); ok {
		return publicKeyFromJWK(key)
	}

	keys, err = refreshJWKS()
	if err != nil {
		return nil, err
	}
	if key, ok := findJWK(keys, kid); ok {
		return publicKeyFromJWK(key)
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

func verifyToken(raw string) (*customClaims, error) {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256"}), // algorithm pinning: the server decides, never the token.
		jwt.WithIssuer(os.Getenv("THUNDER_ISSUER")),
		jwt.WithExpirationRequired(), // a token with no exp claim is rejected, not treated as never-expiring.
		jwt.WithLeeway(30 * time.Second),
	}
	if audience := os.Getenv("THUNDER_AUDIENCE"); audience != "" {
		opts = append(opts, jwt.WithAudience(audience))
	}

	var claims customClaims
	if _, err := jwt.ParseWithClaims(raw, &claims, jwksKeyfunc, opts...); err != nil {
		return nil, err
	}
	return &claims, nil
}

// Gin middleware 

// AuthRequired: AUTHENTICATION. Verifies the JWT and stores the verified
// identity (sub) and permissions (scopes) in the request context.
// Failure → 401: we don't know who you are.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := verifyToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set("sub", claims.Subject)
		// strings.Fields splits on any whitespace and drops empties, so a
		// missing scope claim safely becomes an empty list (deny everything).
		c.Set("scopes", strings.Fields(claims.Scope))
		c.Next()
	}
}

// RequireScope: AUTHORIZATION. Must run AFTER AuthRequired. Checks that the
// verified token carries the given permission.
// Failure → 403: we know who you are, but you may not do this.
func RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, _ := c.Get("scopes")
		scopes, _ := v.([]string)
		for _, s := range scopes {
			if s == scope {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "missing required scope: " + scope,
		})
	}
}

// CurrentSub returns the verified user id inside handlers (or "" if absent).
func CurrentSub(c *gin.Context) string {
	if v, ok := c.Get("sub"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}