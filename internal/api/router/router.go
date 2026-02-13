// Package router defines all API routes and their middleware chains.
// Routes are grouped under /api/v1 with three access levels:
//   - Public: no authentication required (health check, registration)
//   - Authenticated: requires a valid Supabase JWT
//   - Role-restricted: requires authentication + a specific role (e.g., PM)
package router

import (
	"github.com/Ankit1974/TaskDeskBackend/internal/api/handlers"
	"github.com/Ankit1974/TaskDeskBackend/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the Gin router with all API routes.
// It returns the configured engine ready to be started with r.Run().
//
// Route table:
//
//	GET  /api/v1/health     — Public: server and DB health check
//	POST /api/v1/register   — Public: create a new user registration
//	POST /api/v1/projects   — PM only: create a new project
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Apply global middleware here if needed (CORS, Logger, Recovery)

	api := r.Group("/api/v1")
	{
		// ── Public routes (no authentication required) ──
		api.GET("/health", handlers.HealthCheck)
		api.POST("/register", handlers.Register)

		// ── Authenticated routes (valid Supabase JWT required) ──
		auth := api.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// ── PM-only routes (JWT + "PM" role required) ──
			pm := auth.Group("")
			pm.Use(middleware.RequireRole("PM"))
			{
				pm.POST("/projects", handlers.CreateProject)
			}
		}
	}

	return r
}
