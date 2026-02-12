package model

import "time"

type Registration struct {
	ID               string    `json:"id" db:"id"`
	FullName         string    `json:"full_name" db:"full_name" binding:"required"`
	Email            string    `json:"email" db:"email" binding:"required,email"`
	OrganisationName string    `json:"organisation_name" db:"organisation_name"`
	Role             string    `json:"role" db:"role"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
