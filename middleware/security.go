package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const APIKeyHeader = "X-API-Key"

func APIKey(expected string) gin.HandlerFunc {
	expectedDigest := sha256.Sum256([]byte(expected))

	return func(c *gin.Context) {
		provided := c.GetHeader(APIKeyHeader)
		providedDigest := sha256.Sum256([]byte(provided))

		if provided == "" || subtle.ConstantTimeCompare(providedDigest[:], expectedDigest[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
