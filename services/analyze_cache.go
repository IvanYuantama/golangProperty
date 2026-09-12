package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Namespace menggabungkan query dan versi data. Naikkan ANALYZE_CACHE_VERSION setelah data database diperbarui agar cache lama langsung tidak digunakan.
func analysisCacheNamespace(query, version string) string {
	digest := sha256.Sum256([]byte(version + ":" + query))
	return fmt.Sprintf("analyze:%x", digest[:8])
}

func analysisCacheKey(namespace string, latitude, longitude float64) string {
	lat := strconv.FormatFloat(latitude, 'g', -1, 64)
	lng := strconv.FormatFloat(longitude, 'g', -1, 64)
	return fmt.Sprintf("%s:lat:%s:lng:%s", namespace, lat, lng)
}

func (s *AnalyzeService) loadAnalysisCache(ctx context.Context, key string) ([]AnalysisResult, bool) {
	if s.redis == nil {
		return nil, false
	}

	payload, err := s.redis.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		log.Println("cache analisis tidak ditemukan; data akan diambil dari PostgreSQL")
		return nil, false
	}
	if err != nil {
		log.Printf("gagal membaca cache analisis; proses dilanjutkan ke PostgreSQL: %v", err)
		return nil, false
	}

	var results []AnalysisResult
	if err := json.Unmarshal(payload, &results); err != nil {
		log.Printf("cache analisis berisi JSON tidak valid: %v", err)
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			log.Printf("gagal menghapus cache analisis yang rusak: %v", err)
		}
		return nil, false
	}

	log.Println("cache analisis ditemukan; respons dikembalikan tanpa query PostgreSQL")
	return results, true
}

func (s *AnalyzeService) storeAnalysisCache(ctx context.Context, key string, results []AnalysisResult) {
	if s.redis == nil {
		return
	}

	payload, err := json.Marshal(results)
	if err != nil {
		log.Printf("gagal mengubah hasil analisis menjadi JSON untuk cache: %v", err)
		return
	}

	if err := s.redis.Set(ctx, key, payload, s.cacheTTL).Err(); err != nil {
		log.Printf("gagal menyimpan cache analisis: %v", err)
		return
	}

	log.Printf("cache analisis berhasil disimpan dengan masa berlaku %s", s.cacheTTL)
}
