package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthController struct {
	postgres *pgxpool.Pool
	redis    *redis.Client
}

func NewHealthController(postgres *pgxpool.Pool, redisClient *redis.Client) *HealthController {
	return &HealthController{postgres: postgres, redis: redisClient}
}

// Liveness hanya memastikan proses HTTP masih berjalan. Endpoint ini tidak mengakses database sehingga cocok digunakan sebagai healthcheck container.
func (h *HealthController) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readiness memeriksa apakah semua berjalan seperti PostgreSQL wajib tersedia, sedangkan Redis tetap opsional karena API dapat memakai PostgreSQL secara langsung ketika cache bermasalah.
func (h *HealthController) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.postgres.Ping(ctx); err != nil {
		log.Printf("pemeriksaan kesiapan PostgreSQL gagal: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "belum siap",
			"database": "tidak tersedia",
			"redis":    redisStatus(h.redis),
		})
		return
	}

	statusRedis := "dinonaktifkan"
	if h.redis != nil {
		statusRedis = "terhubung"
		if err := h.redis.Ping(ctx).Err(); err != nil {
			log.Printf("peringatan: pemeriksaan kesiapan Redis gagal: %v", err)
			statusRedis = "tidak tersedia; API memakai PostgreSQL langsung"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "siap",
		"database": "terhubung",
		"redis":    statusRedis,
	})
}

func redisStatus(client *redis.Client) string {
	if client == nil {
		return "dinonaktifkan"
	}
	return "belum diperiksa"
}
