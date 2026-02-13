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

// Register handles new user registration.
// It validates the request body, inserts the user into the registrations table,
// and returns the created record with the auto-generated ID and timestamp.
//
// Route: POST /api/v1/register (public, no auth required)
//
// Request body: { full_name, email, organisation_name, role }
// Success response: 201 Created with the full registration record
// Error responses: 400 (validation), 500 (database error)
func Register(c *gin.Context) {
	// Bind and validate the JSON request body against model.Registration rules
	var input model.Registration
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5-second timeout to prevent long-running DB queries from blocking
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the new registration and return the auto-generated ID and created_at
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
