// Package handlers contains all HTTP request handler functions for the API.
// Each handler follows the Gin handler signature: func(c *gin.Context).
// Handlers are responsible for request validation, business logic, DB interaction, and response.
package handlers

import (
	"net/http"

	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/gin-gonic/gin"
)

// HealthCheck returns the server and database health status.
// Used by monitoring tools and load balancers to verify the service is running.
//
// Route: GET /api/v1/health (public, no auth required)
// Response: { "status": "up", "db_status": "up" | "down" }
func HealthCheck(c *gin.Context) {
	// Check database connectivity by pinging the connection pool
	dbStatus := "up"
	if err := db.Pool.Ping(c.Request.Context()); err != nil {
		dbStatus = "down"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "up",
		"db_status": dbStatus,
	})
}
