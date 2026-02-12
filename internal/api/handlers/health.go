package handlers

import (
	"net/http"

	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	dbStatus := "up"
	if err := db.Pool.Ping(c.Request.Context()); err != nil {
		dbStatus = "down"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "up",
		"db_status": dbStatus,
	})
}
