package router

import (
	"github.com/Ankit1974/TaskDeskBackend/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Apply global middleware here if needed (CORS, Logger, Recovery)

	api := r.Group("/api/v1")
	{
		api.GET("/health", handlers.HealthCheck)
		api.POST("/register", handlers.Register)
	}

	return r
}
