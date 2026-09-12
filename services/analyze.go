package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AnalyzeService struct {
	postgres       *pgxpool.Pool
	redis          *redis.Client
	query          string
	cacheNamespace string
	cacheTTL       time.Duration
}

func NewAnalyzeService(
	postgres *pgxpool.Pool,
	redisClient *redis.Client,
	cacheTTL time.Duration,
	cacheVersion string,
) (*AnalyzeService, error) {
	if cacheTTL <= 0 {
		return nil, fmt.Errorf("durasi cache analisis harus lebih besar dari 0")
	}

	query, err := buildAnalyzeQuery(analyzeLayers)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat query analisis: %w", err)
	}

	return &AnalyzeService{
		postgres:       postgres,
		redis:          redisClient,
		query:          query,
		cacheNamespace: analysisCacheNamespace(query, cacheVersion),
		cacheTTL:       cacheTTL,
	}, nil
}

func (s *AnalyzeService) AnalyzeLocation(ctx context.Context, latitude, longitude float64) ([]AnalysisResult, error) {
	// Periksa cacheKey redis dulu baru postgre
	cacheKey := analysisCacheKey(s.cacheNamespace, latitude, longitude)
	if cachedResults, found := s.loadAnalysisCache(ctx, cacheKey); found {
		return cachedResults, nil
	}

	// Nilai koordinat selalu dikirim sebagai parameter SQL, bukan digabungkan ke string query untuk mencegah input pengguna menjadi bagian sintaks SQL.
	rows, err := s.postgres.Query(ctx, s.query, latitude, longitude, analyzeRadiusDegrees)
	if err != nil {
		return nil, fmt.Errorf("gagal menjalankan query analisis spasial: %w", err)
	}
	defer rows.Close()

	results := make([]AnalysisResult, 0, 16)

	for rows.Next() {
		var row analysisRow

		err := rows.Scan(
			&row.Layer,
			&row.StyleValue,
			&row.DistanceMeters,
			&row.Total,
			&row.Attributes,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal membaca baris hasil query: %w", err)
		}

		results = append(results, formatResult(row))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("terjadi kesalahan saat membaca hasil query: %w", err)
	}

	s.storeAnalysisCache(ctx, cacheKey, results)
	return results, nil
}
