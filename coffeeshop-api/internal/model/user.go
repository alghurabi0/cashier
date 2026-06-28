package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents an application user (admin or cashier).
type User struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	TenantID     uuid.UUID `db:"tenant_id"     json:"tenant_id"`
	Username     string    `db:"username"       json:"username"`
	PasswordHash string    `db:"password_hash"  json:"-"`
	Role         string    `db:"role"           json:"role"`
	CreatedAt    time.Time `db:"created_at"     json:"created_at"`
}
