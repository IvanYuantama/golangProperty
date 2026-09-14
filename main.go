package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expressXgolang/golang/config"
	"expressXgolang/golang/controllers"
	"expressXgolang/golang/db"
	appmiddleware "expressXgolang/golang/middleware"
	"expressXgolang/golang/routes"
	"expressXgolang/golang/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// File .env hanya dibaca saat tersedia, misalnya ketika development lokal.
	// Di Docker, environment sudah dimasukkan oleh Compose melalui env_file.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("peringatan: file .env tidak dapat dibaca: %v", err)
	}

	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi aplikasi tidak valid: %v", err)
	}

	// Batas waktu startup mencegah aplikasi menunggu koneksi dependency tanpa batas.
	ctx, cancelStartup := context.WithTimeout(context.Background(), 5*time.Second)

	pool, err := db.OpenPostgres(ctx, appConfig.DatabaseURL)
	if err != nil {
		cancelStartup()
		log.Fatalf("inisialisasi PostgreSQL gagal: %v", err)
	}
	defer pool.Close()

	var redisClient *redis.Client
	if appConfig.RedisURL != "" {
		redisClient, err = db.OpenRedis(ctx, appConfig.RedisURL)
		if err != nil {
			log.Printf("peringatan: Redis tidak tersedia; cache dinonaktifkan: %v", err)
		} else {
			defer redisClient.Close()
			log.Println("koneksi Redis berhasil dibuat")
		}
	}
	cancelStartup()

	analyzeService, err := services.NewAnalyzeService(
		pool,
		redisClient,
		appConfig.AnalyzeCacheTTL,
		appConfig.AnalyzeCacheVersion,
	)
	if err != nil {
		log.Fatalf("inisialisasi layanan analisis gagal: %v", err)
	}
	analyzeController := controllers.NewAnalyzeController(analyzeService)
	healthController := controllers.NewHealthController(pool, redisClient)

	router := gin.New()
	router.HandleMethodNotAllowed = true
	// Urutan middleware dibuat eksplisit agar log, pemulihan panic, dan header keamanan berlaku untuk seluruh endpoint.
	router.Use(
		appmiddleware.RequestLogger(),
		appmiddleware.Recovery(),
		appmiddleware.SecurityHeaders(),
	)

	api := router.Group("/api")
	api.Use(
		appmiddleware.RequestTimeout(appConfig.RequestTimeout),
		appmiddleware.APIKey(appConfig.APIKey),
	)
	routes.Register(router, api, routes.Controllers{
		Analyze: analyzeController,
		Health:  healthController,
	})

	server := &http.Server{
		Addr:              ":" + appConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("server berjalan pada port %s", appConfig.Port)
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server berhenti karena kesalahan: %v", err)
		}
	case <-shutdownSignal.Done():
		log.Println("sinyal penghentian diterima; server sedang dimatikan dengan aman")
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server gagal dimatikan dengan aman: %v", err)
		}
	}
}
