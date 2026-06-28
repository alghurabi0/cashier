package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AdminRepository handles cross-tenant admin queries.
type AdminRepository struct {
	db *sqlx.DB
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// ListTenantsWithCounts returns all tenants with user and device counts.
// Excludes the platform tenant (slug = 'platform').
func (r *AdminRepository) ListTenantsWithCounts() ([]model.TenantWithCounts, error) {
	var tenants []model.TenantWithCounts
	err := r.db.Select(&tenants, `
		SELECT
			t.id, t.name, t.slug, t.is_active, t.settings, t.created_at,
			COALESCE(u.cnt, 0) AS user_count,
			COALESCE(d.cnt, 0) AS device_count
		FROM tenants t
		LEFT JOIN (
			SELECT tenant_id, COUNT(*) AS cnt FROM users GROUP BY tenant_id
		) u ON u.tenant_id = t.id
		LEFT JOIN (
			SELECT tenant_id, COUNT(*) AS cnt FROM devices GROUP BY tenant_id
		) d ON d.tenant_id = t.id
		WHERE t.slug != 'platform'
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// UpdateTenant updates a tenant's mutable fields.
func (r *AdminRepository) UpdateTenant(id uuid.UUID, req model.UpdateTenantRequest) (*model.Tenant, error) {
	// Build dynamic update — only set provided fields
	tenant, err := r.findByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.IsActive != nil {
		tenant.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}

	_, err = r.db.Exec(
		`UPDATE tenants SET name = $1, is_active = $2, settings = $3 WHERE id = $4`,
		tenant.Name, tenant.IsActive, tenant.Settings, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (r *AdminRepository) findByID(id uuid.UUID) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Get(&tenant,
		`SELECT id, name, slug, is_active, settings, created_at
		 FROM tenants WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return &tenant, nil
}

// GetPlatformStats returns platform-wide statistics.
func (r *AdminRepository) GetPlatformStats() (*model.PlatformStats, error) {
	var stats model.PlatformStats

	// Tenants (exclude platform)
	r.db.Get(&stats.TotalTenants,
		`SELECT COUNT(*) FROM tenants WHERE slug != 'platform'`)
	r.db.Get(&stats.ActiveTenants,
		`SELECT COUNT(*) FROM tenants WHERE slug != 'platform' AND is_active = true`)

	// Users (exclude platform tenant users)
	r.db.Get(&stats.TotalUsers,
		`SELECT COUNT(*) FROM users WHERE tenant_id != '00000000-0000-0000-0000-000000000001'`)

	// Devices
	r.db.Get(&stats.TotalDevices,
		`SELECT COUNT(*) FROM devices`)

	// Orders
	r.db.Get(&stats.TotalOrders,
		`SELECT COUNT(*) FROM orders`)
	r.db.Get(&stats.TodayOrders,
		`SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE`)

	return &stats, nil
}
