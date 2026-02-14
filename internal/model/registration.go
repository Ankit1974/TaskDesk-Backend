// Package model defines the data structures used across the application.
// These structs serve dual purpose: JSON serialization for API requests/responses
// and database row mapping via struct tags (`json`, `db`, `binding`).
package model

import "time"

/*
Registration represents a user registration record.
Used both as the request body for POST /api/v1/register and as the DB row struct.

Struct tags:
  - `json`:    JSON field name for API request/response serialization
  - `db`:      Database column name for row scanning
  - `binding`: Gin validation rules (e.g., "required", "email")
*/
type Registration struct {
	ID               string    `json:"id" db:"id"`
	FullName         string    `json:"full_name" db:"full_name" binding:"required"`
	Email            string    `json:"email" db:"email" binding:"required,email"`
	OrganisationName string    `json:"organisation_name" db:"organisation_name" binding:"required"`
	Role             string    `json:"role" db:"role" binding:"required"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
