package model

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleViewers Role = "viewers"
	RoleAnalyst Role = "analyst"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"cr3eated_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
