package model

import (
	"time"

	"github.com/google/uuid"
)

// Table represents a physical table in the coffeeshop.
// Each table has a unique token used for QR-code web menu access.
type Table struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	Number    string    `db:"number"     json:"number"`
	Token     string    `db:"token"      json:"token"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateTableRequest is the body for POST /api/v1/tables.
type CreateTableRequest struct {
	Number string `json:"number"` // e.g. "1", "A1"
}
