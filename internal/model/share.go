package model

import (
	"time"

	"github.com/google/uuid"
)

type SharedAccess struct {
	ID           uuid.UUID `json:"id"`
	OwnerID      uuid.UUID `json:"owner_id"`
	SharedWithID uuid.UUID `json:"shared_with_id"`
	Permission   string    `json:"permission"`
	CreatedAt    time.Time `json:"created_at"`
}

type ShareRequest struct {
	SharedWithEmail string `json:"shared_with_email"`
	Permission      string `json:"permission"` // "viewer" or "analyst"
}
