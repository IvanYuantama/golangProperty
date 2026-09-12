package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type App struct {
	Port                string
	DatabaseURL         string
	RedisURL            string
	APIKey              string
	EnableHSTS          bool
	TrustedProxies      []string
	RateLimitPerMinute  int
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
		TrustedProxies:      commaSeparatedEnv("TRUSTED_PROXIES"),
		RateLimitPerMinute:  60,
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

	if raw := strings.TrimSpace(os.Getenv("ENABLE_HSTS")); raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return App{}, fmt.Errorf("ENABLE_HSTS harus bernilai true atau false")
		}
		app.EnableHSTS = enabled
	}

	if raw := strings.TrimSpace(os.Getenv("RATE_LIMIT_PER_MINUTE")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return App{}, fmt.Errorf("RATE_LIMIT_PER_MINUTE harus berupa angka lebih besar dari 0")
		}
		app.RateLimitPerMinute = limit
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

func commaSeparatedEnv(name string) []string {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil
	}

	values := strings.Split(raw, ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
