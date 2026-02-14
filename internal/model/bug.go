package model

import "time"

// CreateBugRequest represents a single bug in the batch creation request.
type CreateBugRequest struct {
	Title       string   `json:"title" binding:"required"`
	Priority    string   `json:"priority" binding:"required,oneof=critical high medium low"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	Version     string   `json:"version"`
	Platform    string   `json:"platform"`
	AssignedTo  string   `json:"assigned_to"` // Optional registration UUID
}

// CreateBugsRequest wraps an array of bugs for POST /api/v1/projects/:id/bugs.
type CreateBugsRequest struct {
	Bugs []CreateBugRequest `json:"bugs" binding:"required,min=1,max=20,dive"`
}

// Bug represents a full bug record in the database.
type Bug struct {
	ID          string    `json:"id" db:"id"`
	ProjectID   string    `json:"project_id" db:"project_id"`
	BugNumber   string    `json:"bug_number" db:"bug_number"`
	Title       string    `json:"title" db:"title"`
	Priority    string    `json:"priority" db:"priority"`
	Description *string   `json:"description" db:"description"`
	Steps       []string  `json:"steps" db:"steps"`
	Version     *string   `json:"version,omitempty" db:"version"`
	Platform    *string   `json:"platform,omitempty" db:"platform"`
	Status      string    `json:"status" db:"status"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	AssignedTo  *string   `json:"assigned_to" db:"assigned_to"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateBugsResponse wraps the created bugs array for the API response.
type CreateBugsResponse struct {
	Bugs  []Bug `json:"bugs"`
	Count int   `json:"count"`
}
