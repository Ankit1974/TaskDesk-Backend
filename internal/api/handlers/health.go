// Package handlers contains all HTTP request handler functions for the API.
// Each handler follows the Gin handler signature: func(c *gin.Context).
package handlers

import (
	"net/http"

	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/gin-gonic/gin"
)

// HealthCheck returns the server and database health status.
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
