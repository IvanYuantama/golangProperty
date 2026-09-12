package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const APIKeyHeader = "X-API-Key"

// APIKey melindungi route dengan API key yang dikirim melalui header X-API-Key.
// Digest berukuran tetap dibandingkan secara constant-time agar nilai rahasia
// tidak dibandingkan menggunakan operasi string biasa.
func APIKey(expected string) gin.HandlerFunc {
	expectedDigest := sha256.Sum256([]byte(expected))

	return func(c *gin.Context) {
		provided := c.GetHeader(APIKeyHeader)
		providedDigest := sha256.Sum256([]byte(provided))

		if provided == "" || subtle.ConstantTimeCompare(providedDigest[:], expectedDigest[:]) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "API key tidak valid atau tidak dikirim",
			})
			return
		}

		c.Next()
	}
}

// SecurityHeaders menambahkan header defensif untuk respons API.
func SecurityHeaders(enableHSTS bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "no-store")
		if enableHSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

type clientWindow struct {
	requests int
	resetAt  time.Time
}

// IPRateLimiter adalah fixed-window limiter untuk satu instance aplikasi.
// Untuk beberapa instance backend, limiter sebaiknya dipindahkan ke Redis.
type IPRateLimiter struct {
	mu          sync.Mutex
	clients     map[string]clientWindow
	limit       int
	window      time.Duration
	lastCleanup time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		clients:     make(map[string]clientWindow),
		limit:       limit,
		window:      window,
		lastCleanup: time.Now(),
	}
}

func (l *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		clientIP := c.ClientIP()

		l.mu.Lock()
		if now.Sub(l.lastCleanup) >= l.window {
			for ip, window := range l.clients {
				if !now.Before(window.resetAt) {
					delete(l.clients, ip)
				}
			}
			l.lastCleanup = now
		}

		window, exists := l.clients[clientIP]
		if !exists || !now.Before(window.resetAt) {
			window = clientWindow{resetAt: now.Add(l.window)}
		}

		if window.requests >= l.limit {
			retryAfter := int(math.Ceil(time.Until(window.resetAt).Seconds()))
			if retryAfter < 1 {
				retryAfter = 1
			}
			l.mu.Unlock()

			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "batas permintaan terlampaui; silakan coba lagi nanti",
			})
			return
		}

		window.requests++
		l.clients[clientIP] = window
		l.mu.Unlock()

		c.Next()
	}
}
