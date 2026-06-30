package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a customer/coffee shop in the SaaS platform.
type Tenant struct {
	ID                     uuid.UUID      `db:"id"                        json:"id"`
	Name                   string         `db:"name"                      json:"name"`
	Slug                   string         `db:"slug"                      json:"slug"`
	IsActive               bool           `db:"is_active"                 json:"is_active"`
	Settings               TenantSettings `db:"settings"                  json:"settings"`
	ProvisionCode          *string        `db:"provision_code"            json:"provision_code,omitempty"`
	ProvisionCodeExpiresAt *time.Time     `db:"provision_code_expires_at" json:"provision_code_expires_at,omitempty"`
	CreatedAt              time.Time      `db:"created_at"                json:"created_at"`
}

// TenantSettings holds tenant-level configuration (stored as JSONB).
type TenantSettings struct {
	KitchenModeEnabled     bool   `json:"kitchen_mode_enabled"`
	ConflictResolutionMode string `json:"conflict_resolution_mode"` // "last-write-wins" | "manual"
	MenuURL                string `json:"menu_url"`                 // where this tenant's web menu is hosted
	IntroVideoURL          string `json:"intro_video_url"`          // R2 URL for the POS login background video
}

// Scan implements sql.Scanner for reading JSONB from the database.
func (s *TenantSettings) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("unsupported type for TenantSettings: %T", src)
	}
}

// Value implements driver.Valuer for writing JSONB to the database.
func (s TenantSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// CreateTenantRequest is the body for POST /api/v1/tenants (self-service signup).
type CreateTenantRequest struct {
	Name          string `json:"name"`           // "NJ Coffee"
	Slug          string `json:"slug"`           // "nj-coffee"
	AdminUsername string `json:"admin_username"` // first admin user
	AdminPassword string `json:"admin_password"`
}

// CreateTenantResponse is returned after successful tenant creation.
type CreateTenantResponse struct {
	Tenant *Tenant `json:"tenant"`
	Token  string  `json:"token"`
	User   *User   `json:"user"`
}

// ProvisionRequest is the body for POST /api/v1/provision.
type ProvisionRequest struct {
	Code string `json:"code"`
}

// ProvisionResponse is returned after successful provisioning.
type ProvisionResponse struct {
	Tenant   *Tenant `json:"tenant"`
	Token    string  `json:"token"`
	Username string  `json:"username"`
	Password string  `json:"password"`
}
