package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type App struct {
	Port                string
	DatabaseURL         string
	RedisURL            string
	APIKey              string
	RequestTimeout      time.Duration
	AnalyzeCacheTTL     time.Duration
	AnalyzeCacheVersion string
}

func Load() (App, error) {
	app := App{
		Port:                strings.TrimSpace(os.Getenv("PORT")),
		DatabaseURL:         strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:            strings.TrimSpace(os.Getenv("REDIS_URL")),
		APIKey:              strings.TrimSpace(os.Getenv("API_KEY")),
		RequestTimeout:      5 * time.Second,
		AnalyzeCacheTTL:     10 * time.Minute,
		AnalyzeCacheVersion: strings.TrimSpace(os.Getenv("ANALYZE_CACHE_VERSION")),
	}
	if app.Port == "" {
		app.Port = "3000"
	}

	if app.DatabaseURL == "" {
		return App{}, fmt.Errorf("DATABASE_URL wajib diisi")
	}
	if len(app.APIKey) < 32 {
		return App{}, fmt.Errorf("API_KEY wajib diisi dengan nilai acak minimal 32 karakter")
	}
	if app.AnalyzeCacheVersion == "" {
		app.AnalyzeCacheVersion = "v1"
	}

	if raw := strings.TrimSpace(os.Getenv("REQUEST_TIMEOUT")); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil || timeout <= 0 {
			return App{}, fmt.Errorf("REQUEST_TIMEOUT harus berupa durasi lebih besar dari 0, misalnya 5s")
		}
		app.RequestTimeout = timeout
	}

	if raw := strings.TrimSpace(os.Getenv("ANALYZE_CACHE_TTL")); raw != "" {
		ttl, err := time.ParseDuration(raw)
		if err != nil || ttl <= 0 {
			return App{}, fmt.Errorf("ANALYZE_CACHE_TTL harus berupa durasi lebih besar dari 0, misalnya 10m")
		}
		app.AnalyzeCacheTTL = ttl
	}

	return app, nil
}
