package mcpserver

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
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
)

type mcpJWTHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type mcpJWTClaims struct {
	Iss   string          `json:"iss"`
	Sub   string          `json:"sub"`
	Exp   int64           `json:"exp"`
	Scope string          `json:"scope"`
	Aud   json.RawMessage `json:"aud"`
}

func (c *mcpJWTClaims) audience() []string {
	if len(c.Aud) == 0 {
		return nil
	}
	var single string
	if json.Unmarshal(c.Aud, &single) == nil {
		return []string{single}
	}
	var many []string
	if json.Unmarshal(c.Aud, &many) == nil {
		return many
	}
	return nil
}

type mcpParsedJWT struct {
	header       mcpJWTHeader
	claims       mcpJWTClaims
	signingInput string
	signature    []byte
}

func parseMCPJWT(raw string) (*mcpParsedJWT, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("bad token encoding")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("bad token encoding")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("bad token encoding")
	}
	var p mcpParsedJWT
	if json.Unmarshal(headerBytes, &p.header) != nil || json.Unmarshal(payloadBytes, &p.claims) != nil {
		return nil, errors.New("bad token JSON")
	}
	p.signingInput = parts[0] + "." + parts[1]
	p.signature = sig
	return &p, nil
}

type mcpJWK struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	mcpJWKSMu    sync.Mutex
	mcpJWKSCache []mcpJWK
)

func mcpHTTPClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	if os.Getenv("THUNDER_INSECURE_TLS") == "true" {
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return client
}

func getMCPJWKS() ([]mcpJWK, error) {
	mcpJWKSMu.Lock()
	defer mcpJWKSMu.Unlock()
	if mcpJWKSCache != nil {
		return mcpJWKSCache, nil
	}
	resp, err := mcpHTTPClient().Get(thunderIssuerURL() + "/oauth2/jwks")
	if err != nil {
		return nil, errors.New("cannot reach ThunderID JWKS")
	}
	defer resp.Body.Close()
	var doc struct {
		Keys []mcpJWK `json:"keys"`
	}
	if json.NewDecoder(resp.Body).Decode(&doc) != nil {
		return nil, errors.New("bad JWKS document")
	}
	mcpJWKSCache = doc.Keys
	return mcpJWKSCache, nil
}

func mcpPublicKeyFromJWK(k mcpJWK) (*rsa.PublicKey, error) {
	nBytes, err1 := base64.RawURLEncoding.DecodeString(k.N)
	eBytes, err2 := base64.RawURLEncoding.DecodeString(k.E)
	if err1 != nil || err2 != nil || len(eBytes) > 8 {
		return nil, errors.New("bad JWK")
	}
	padded := make([]byte, 8)
	copy(padded[8-len(eBytes):], eBytes)
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(binary.BigEndian.Uint64(padded))}, nil
}

// verifyMCPToken runs the full check: RS256 signature against ThunderID's
// JWKS, issuer match, expiry (30s clock skew), and audience match against
// this MCP server's own resource identifier (MCP_RESOURCE_ID) - NOT the
// REST API's THUNDER_AUDIENCE, since this is a different resource server.
func verifyMCPToken(raw string) (sub string, scopes []string, err error) {
	p, err := parseMCPJWT(raw)
	if err != nil {
		return "", nil, err
	}
	if p.header.Alg != "RS256" {
		return "", nil, errors.New("unsupported token algorithm")
	}
	keys, err := getMCPJWKS()
	if err != nil {
		return "", nil, err
	}
	var key *rsa.PublicKey
	for _, k := range keys {
		if k.Kid == p.header.Kid {
			if key, err = mcpPublicKeyFromJWK(k); err != nil {
				return "", nil, err
			}
			break
		}
	}
	if key == nil {
		return "", nil, errors.New("signing key not found")
	}
	digest := sha256.Sum256([]byte(p.signingInput))
	if rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], p.signature) != nil {
		return "", nil, errors.New("invalid token signature")
	}
	if p.claims.Iss != thunderIssuerURL() {
		return "", nil, errors.New("invalid token issuer")
	}
	if p.claims.Exp != 0 && p.claims.Exp < time.Now().Unix()-30 {
		return "", nil, errors.New("token has expired")
	}
	expected := mcpResourceID()
	found := false
	for _, a := range p.claims.audience() {
		if a == expected {
			found = true
		}
	}
	if !found {
		return "", nil, errors.New("invalid token audience")
	}
	return p.claims.Sub, strings.Fields(p.claims.Scope), nil
}