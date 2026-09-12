package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenPostgres(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}

	if !strings.Contains(connStr, "sslmode=") {
		delimiter := "?"
		if strings.Contains(connStr, "?") {
			delimiter = "&"
		}
		connStr = fmt.Sprintf("%s%ssslmode=require", connStr, delimiter)
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("format koneksi PostgreSQL tidak valid: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pool koneksi PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal menghubungi PostgreSQL: %w", err)
	}

	log.Println("koneksi PostgreSQL berhasil dibuat")
	return pool, nil
}
