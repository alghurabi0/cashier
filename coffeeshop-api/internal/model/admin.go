package model

import (
	"time"

	"github.com/google/uuid"
)

// UpdateTenantRequest is the body for PUT /api/v1/admin/tenants/{id}.
type UpdateTenantRequest struct {
	Name     *string         `json:"name,omitempty"`
	IsActive *bool           `json:"is_active,omitempty"`
	Settings *TenantSettings `json:"settings,omitempty"`
}

// CreateUserRequest is the body for POST /api/v1/admin/tenants/{id}/users.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // "admin" or "cashier"
}

// TenantWithCounts extends Tenant with user/device counts for admin listing.
type TenantWithCounts struct {
	ID         uuid.UUID      `db:"id"          json:"id"`
	Name       string         `db:"name"        json:"name"`
	Slug       string         `db:"slug"        json:"slug"`
	IsActive   bool           `db:"is_active"   json:"is_active"`
	Settings   TenantSettings `db:"settings"    json:"settings"`
	CreatedAt  time.Time      `db:"created_at"  json:"created_at"`
	UserCount  int            `db:"user_count"  json:"user_count"`
	DeviceCount int           `db:"device_count" json:"device_count"`
}

// TenantDetail is the full detail view for a single tenant.
type TenantDetail struct {
	Tenant  *Tenant  `json:"tenant"`
	Users   []User   `json:"users"`
	Devices []Device `json:"devices"`
}

// PlatformStats contains platform-wide statistics for the admin dashboard.
type PlatformStats struct {
	TotalTenants  int `json:"total_tenants"`
	ActiveTenants int `json:"active_tenants"`
	TotalUsers    int `json:"total_users"`
	TotalDevices  int `json:"total_devices"`
	TotalOrders   int `json:"total_orders"`
	TodayOrders   int `json:"today_orders"`
}
