package model

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a registered POS terminal or display under a tenant.
type Device struct {
	ID         uuid.UUID  `db:"id"           json:"id"`
	TenantID   uuid.UUID  `db:"tenant_id"    json:"tenant_id"`
	DeviceName string     `db:"device_name"  json:"device_name"`
	DeviceType string     `db:"device_type"  json:"device_type"` // "pos", "kitchen_display"
	IsActive   bool       `db:"is_active"    json:"is_active"`
	LastSeenAt *time.Time `db:"last_seen_at" json:"last_seen_at"`
	CreatedAt  time.Time  `db:"created_at"   json:"created_at"`
}

// RegisterDeviceRequest is the body for POST /api/v1/devices/register.
type RegisterDeviceRequest struct {
	DeviceName string `json:"device_name"` // "POS-Counter-1"
	DeviceType string `json:"device_type"` // "pos" (default), "kitchen_display"
}
