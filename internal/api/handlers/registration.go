package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/Ankit1974/TaskDeskBackend/internal/model"
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var input model.Registration
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO registrations (full_name, email, organisation_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := db.Pool.QueryRow(ctx, query,
		input.FullName,
		input.Email,
		input.OrganisationName,
		input.Role,
	).Scan(&input.ID, &input.CreatedAt)

	if err != nil {
		logger.Log.Error("Failed to insert registration: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save registration"})
		return
	}

	c.JSON(http.StatusCreated, input)
}
