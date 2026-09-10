package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"expressXgolang/golang/controllers"
	"expressXgolang/golang/db"
	"expressXgolang/golang/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file was not loaded: %v", err)
	}

	ctx := context.Background()

	pool, err := db.InitDB()
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer pool.Close()

	redisClient, err := db.InitRedis(ctx)
	if err != nil {
		log.Printf("warning: Redis unavailable; cache disabled: %v", err)
	} else {
		defer redisClient.Close()
		log.Println("Redis connection established successfully")
	}

	router := gin.Default()
	router.GET("/api/analyze", controllers.GetAnalyzeLocation)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("server running on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// runAnalyzeExample mempertahankan contoh pemanggilan service langsung Anda.
// Fungsi ini tidak dipanggil ketika HTTP server dijalankan.
func runAnalyzeExample() {
	results, err := services.AnalyzeLocation(
		context.Background(),
		-8.65,
		115.2167,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", results)
}
