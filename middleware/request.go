package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Kalo server backend kelamaan ngasihnya ntar dikasih timeout
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RequestLogger mencatat masalah di log VPS lebih mudah ditelusuri.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if c.Request.URL.Path == "/health" && c.Writer.Status() < http.StatusBadRequest {
			return
		}

		log.Printf(
			"permintaan selesai: metode=%s path=%s status=%d durasi=%s ip=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).Round(time.Microsecond),
			c.ClientIP(),
		)
	}
}

// Recovery menangkap panic agar satu request bermasalah tidak mematikan server.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic ditangani: metode=%s path=%s penyebab=%v", c.Request.Method, c.Request.URL.Path, recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "terjadi kesalahan internal pada server",
		})
	})
}
