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

// GetProjects returns all projects created by or assigned to the authenticated user.
// Supports optional query parameters:
//   - status: filter by project status (active, planning, on_hold, completed)
//   - search: case-insensitive search on project_name
//   - page:   page number (default: 1)
//   - limit:  items per page (default: 10, max: 50)
func GetProjects(c *gin.Context) {
	// Get the authenticated user (set by AuthMiddleware)
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse and validate pagination query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Validate optional status filter against allowed values
	status := c.Query("status")
	var statusParam *string
	if status != "" {
		validStatuses := map[string]bool{
			"active": true, "planning": true, "on_hold": true, "completed": true,
		}
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status. Must be one of: active, planning, on_hold, completed",
			})
			return
		}
		statusParam = &status
	}

	// Optional search filter for project name
	search := c.Query("search")
	var searchParam *string
	if search != "" {
		searchParam = &search
	}

	// 5-second timeout for database queries
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count total matching projects (for pagination metadata)
	countQuery := `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		WHERE p.id IN (
			SELECT id FROM projects WHERE created_by = $1
			UNION
			SELECT project_id FROM project_members WHERE user_id = $1
		)
		AND ($2::VARCHAR IS NULL OR p.status = $2)
		AND ($3::VARCHAR IS NULL OR p.project_name ILIKE '%' || $3 || '%')
	`

	var totalCount int
	err := db.Pool.QueryRow(ctx, countQuery,
		user.RegistrationID, statusParam, searchParam,
	).Scan(&totalCount)
	if err != nil {
		logger.Log.Error("Failed to count projects: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// Fetch the paginated project list ordered by most recently updated
	dataQuery := `
		SELECT p.id, p.project_name, p.description, p.icon, p.teams,
		       p.start_date, p.status, p.workspace_id, p.created_by,
		       p.progress, p.member_count, p.created_at, p.updated_at
		FROM projects p
		WHERE p.id IN (
			SELECT id FROM projects WHERE created_by = $1
			UNION
			SELECT project_id FROM project_members WHERE user_id = $1
		)
		AND ($2::VARCHAR IS NULL OR p.status = $2)
		AND ($3::VARCHAR IS NULL OR p.project_name ILIKE '%' || $3 || '%')
		ORDER BY p.updated_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := db.Pool.Query(ctx, dataQuery,
		user.RegistrationID, statusParam, searchParam, limit, offset,
	)
	if err != nil {
		logger.Log.Error("Failed to query projects: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	defer rows.Close()

	// Scan rows into Project structs (empty slice, not nil, for clean JSON [])
	projects := []model.Project{}
	for rows.Next() {
		var p model.Project
		var startDate *time.Time

		err := rows.Scan(
			&p.ID, &p.ProjectName, &p.Description, &p.Icon, &p.Teams,
			&startDate, &p.Status, &p.WorkspaceID, &p.CreatedBy,
			&p.Progress, &p.MemberCount, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			logger.Log.Error("Failed to scan project row: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
			return
		}

		// Convert *time.Time to *string for the response (YYYY-MM-DD format)
		if startDate != nil {
			formatted := startDate.Format("2006-01-02")
			p.StartDate = &formatted
		}
		if p.Teams == nil {
			p.Teams = []string{}
		}

		projects = append(projects, p)
	}

	if rows.Err() != nil {
		logger.Log.Error("Row iteration error: " + rows.Err().Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, model.ProjectListResponse{
		Projects:   projects,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	})
}

// GetProjectByID returns the full details of a single project.
// The user must be the project creator or an assigned member to view it.
func GetProjectByID(c *gin.Context) {
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

	// 5-second timeout for the database query
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch the project only if the user is the creator or an assigned member
	query := `
		SELECT p.id, p.project_name, p.description, p.icon, p.teams,
		       p.start_date, p.status, p.workspace_id, p.created_by,
		       p.progress, p.member_count, p.created_at, p.updated_at
		FROM projects p
		WHERE p.id = $1
		AND p.id IN (
			SELECT id FROM projects WHERE created_by = $2
			UNION
			SELECT project_id FROM project_members WHERE user_id = $2
		)
	`

	var project model.Project
	var startDate *time.Time

	err := db.Pool.QueryRow(ctx, query, projectID, user.RegistrationID).Scan(
		&project.ID, &project.ProjectName, &project.Description, &project.Icon, &project.Teams,
		&startDate, &project.Status, &project.WorkspaceID, &project.CreatedBy,
		&project.Progress, &project.MemberCount, &project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		logger.Log.Error("Failed to fetch project: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	// Convert *time.Time to *string for the response (YYYY-MM-DD format)
	if startDate != nil {
		formatted := startDate.Format("2006-01-02")
		project.StartDate = &formatted
	}
	if project.Teams == nil {
		project.Teams = []string{}
	}

	c.JSON(http.StatusOK, project)
}
