package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func OpenRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL wajib diisi")
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("format REDIS_URL tidak valid")
	}
	options.DialTimeout = 2 * time.Second
	options.ReadTimeout = 750 * time.Millisecond
	options.WriteTimeout = 750 * time.Millisecond
	options.MaxRetries = 1

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("gagal menghubungi Redis: %w", err)
	}

	return client, nil
}
