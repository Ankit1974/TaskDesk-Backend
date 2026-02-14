package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/api/middleware"
	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/Ankit1974/TaskDeskBackend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// CreateBugs handles batch bug creation for a specific project.
// Only the project creator or assigned members can create bugs.
// Accepts 1–20 bugs per request.
func CreateBugs(c *gin.Context) {
	// Get the authenticated user (set by AuthMiddleware)
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	// 5-second timeout for all database operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify the user has access to this project (creator or member)
	accessQuery := `
		SELECT EXISTS(
			SELECT 1 FROM projects WHERE id = $1 AND created_by = $2
			UNION
			SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2
		)
	`
	var hasAccess bool
	err := db.Pool.QueryRow(ctx, accessQuery, projectID, user.RegistrationID).Scan(&hasAccess)
	if err != nil {
		logger.Log.Error("Failed to check project access: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify project access"})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this project"})
		return
	}

	// Bind and validate the JSON request body
	var input model.CreateBugsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the current max bug number for this project to generate sequential IDs
	var currentMax int
	err = db.Pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(CAST(SUBSTRING(bug_number FROM 5) AS INTEGER)), 0) FROM bugs WHERE project_id = $1`,
		projectID,
	).Scan(&currentMax)
	if err != nil {
		logger.Log.Error("Failed to get max bug number: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bugs"})
		return
	}

	// Batch insert all bugs using pgx.Batch for efficiency
	batch := &pgx.Batch{}
	insertQuery := `
		INSERT INTO bugs (project_id, bug_number, title, priority, description, steps, version, platform, created_by, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, project_id, bug_number, title, priority, description, steps, version, platform, status, created_by, assigned_to, created_at, updated_at
	`

	for i, bug := range input.Bugs {
		bugNumber := fmt.Sprintf("BUG-%d", currentMax+i+1)

		// Convert optional fields: empty string → nil for SQL NULL
		var description *string
		if bug.Description != "" {
			description = &bug.Description
		}

		var version *string
		if bug.Version != "" {
			version = &bug.Version
		}

		var platform *string
		if bug.Platform != "" {
			platform = &bug.Platform
		}

		var assignedTo *string
		if bug.AssignedTo != "" {
			assignedTo = &bug.AssignedTo
		}

		steps := bug.Steps
		if steps == nil {
			steps = []string{}
		}

		batch.Queue(insertQuery,
			projectID, bugNumber, bug.Title, bug.Priority,
			description, steps, version, platform,
			user.RegistrationID, assignedTo,
		)
	}

	// Execute the batch
	br := db.Pool.SendBatch(ctx, batch)
	defer br.Close()

	// Collect the returned bugs
	bugs := make([]model.Bug, 0, len(input.Bugs))
	for range input.Bugs {
		var b model.Bug
		err := br.QueryRow().Scan(
			&b.ID, &b.ProjectID, &b.BugNumber, &b.Title, &b.Priority,
			&b.Description, &b.Steps, &b.Version, &b.Platform,
			&b.Status, &b.CreatedBy, &b.AssignedTo,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			logger.Log.Error("Failed to insert bug: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bugs"})
			return
		}
		if b.Steps == nil {
			b.Steps = []string{}
		}
		bugs = append(bugs, b)
	}

	c.JSON(http.StatusCreated, model.CreateBugsResponse{
		Bugs:  bugs,
		Count: len(bugs),
	})
}
