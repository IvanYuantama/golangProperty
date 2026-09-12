package routes

import (
	"expressXgolang/golang/controllers"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	Analyze *controllers.AnalyzeController
	Health  *controllers.HealthController
}

// Buat daftarin endpoint baru
// Endpoint health tidak memakai API key agar Docker/reverse proxy dapat memeriksa kondisi aplikasi. Endpoint app tetap berada di grup /api.
func Register(router *gin.Engine, api *gin.RouterGroup, handlers Controllers) {
	router.NoRoute(controllers.NotFound)
	router.NoMethod(controllers.MethodNotAllowed)
	router.GET("/health", handlers.Health.Liveness)
	router.GET("/ready", handlers.Health.Readiness)
	api.GET("/analyze", handlers.Analyze.GetLocation)
}
