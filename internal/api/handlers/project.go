package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/api/middleware"
	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/Ankit1974/TaskDeskBackend/internal/model"
	"github.com/gin-gonic/gin"
)

// generateWorkspaceID creates a short, human-readable workspace identifier.
func generateWorkspaceID(projectName string) string {
	name := strings.TrimSpace(projectName)
	prefix := "UNK"
	if len(name) >= 3 {
		prefix = strings.ToUpper(name[:3])
	} else if len(name) > 0 {
		prefix = strings.ToUpper(name)
	}

	// Convert current timestamp (milliseconds) to base36 and take the last 4 characters
	suffix := strings.ToUpper(strconv.FormatInt(time.Now().UnixMilli(), 36))
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}

	return fmt.Sprintf("%s-%s", prefix, suffix)
}

// CreateProject handles project creation. Only accessible by users with the "PM" role.
// Error responses: 400 (validation), 401 (unauthenticated), 403 (not PM), 500 (database error)
func CreateProject(c *gin.Context) {
	// Get the authenticated user (set by AuthMiddleware)
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Bind and validate the JSON request body against model.CreateProjectRequest rules
	var input model.CreateProjectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a unique workspace identifier for the project
	workspaceID := generateWorkspaceID(input.ProjectName)

	// Parse the optional start_date
	var startDate *time.Time
	if input.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
		startDate = &parsed
	}

	// 5-second timeout for the database insert operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the project with default status "planning", progress 0, member_count 0.
	query := `
		INSERT INTO projects (project_name, description, icon, teams, start_date, status, workspace_id, created_by)
		VALUES ($1, $2, $3, $4, $5, 'planning', $6, $7)
		RETURNING id, status, progress, member_count, created_at, updated_at
	`

	// Pre-populate the response struct with client-provided values
	var project model.Project
	project.ProjectName = input.ProjectName
	project.Description = input.Description
	//project.Icon = input.Icon
	project.Teams = input.Teams
	if project.Teams == nil {
		project.Teams = []string{} // Ensure JSON serializes as [] instead of null
	}
	project.WorkspaceID = workspaceID
	project.CreatedBy = user.RegistrationID

	// Execute the insert and scan the returned auto-generated fields
	err := db.Pool.QueryRow(ctx, query,
		input.ProjectName,
		input.Description,
		//input.Icon,
		input.Teams, // pgx natively converts []string to PostgreSQL TEXT[]
		startDate,   // nil becomes SQL NULL for optional dates
		workspaceID,
		user.RegistrationID, // The PM's registration UUID
	).Scan(
		&project.ID,
		&project.Status,
		&project.Progress,
		&project.MemberCount,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		logger.Log.Error("Failed to create project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	// Set the start_date string in the response (only if provided by the client)
	if input.StartDate != "" {
		project.StartDate = &input.StartDate
	}

	c.JSON(http.StatusCreated, project)
}
