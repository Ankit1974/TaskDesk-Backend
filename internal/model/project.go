package model

import "time"

/*
	CreateProjectRequest represents the JSON body for POST /api/v1/projects.
	Only the fields the client sends are included here (no server-generated fields).
*/
/*
	Validation rules (via Gin's binding tags):
	  - project_name: required
	  - description:  required
	  - icon:         required, must be one of the allowed Material Icon names
	  - teams:        required, each entry must be a valid team key
	  - start_date:   optional, expected format: "YYYY-MM-DD"
*/
type CreateProjectRequest struct {
	ProjectName string `json:"project_name" binding:"required"`
	Description string `json:"description" binding:"required"`
	// Icon        string   `json:"icon" binding:"required,oneof=language smartphone cloud storage cloud-upload"`
	Teams     []string `json:"teams" binding:"required,dive,oneof=backend frontend mobile qa uiux"`
	StartDate string   `json:"start_date" binding:"omitempty"`
}

/*
Project represents a full project record in the database.
Used as the API response after project creation.

Server-generated fields (not sent by the client):
  - ID, Status, WorkspaceID, CreatedBy, Progress, MemberCount, CreatedAt, UpdatedAt
*/
type Project struct {
	ID          string    `json:"id" db:"id"`
	ProjectName string    `json:"project_name" db:"project_name"`
	Description string    `json:"description" db:"description"`
	Icon        string    `json:"icon" db:"icon"`
	Teams       []string  `json:"teams" db:"teams"`
	StartDate   *string   `json:"start_date,omitempty" db:"start_date"`
	Status      string    `json:"status" db:"status"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	Progress    int       `json:"progress" db:"progress"`
	MemberCount int       `json:"member_count" db:"member_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
