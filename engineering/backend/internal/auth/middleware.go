package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Gin middleware

// AuthRequired: AUTHENTICATION. Verifies the JWT and stores the verified
// identity (sub) and permissions (scopes) in the request context.
// Failure → 401: we don't know who you are.
func AuthRequired(v *Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		claims, err := v.VerifyToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.Set("sub", claims.Subject)
		// strings.Fields splits on any whitespace and drops empties, so a
		// missing scope claim safely becomes an empty list (deny everything).
		c.Set("scopes", strings.Fields(claims.Scope))
		c.Next()
	}
}

// respondAuthError logs the real verification error and returns a
// generic message — never the raw error — distinguishing a bad/expired
// token (401) from a failure to reach ThunderID (503).
func respondAuthError(c *gin.Context, err error) {
	log.Printf("token verification failed: %v", err)
	if errors.Is(err, ErrJWKSUnavailable) {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "authentication temporarily unavailable"})
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
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
