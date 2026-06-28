package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TenantRepository handles database operations for tenants.
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create inserts a new tenant.
func (r *TenantRepository) Create(name, slug string, settings model.TenantSettings) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Get(&tenant,
		`INSERT INTO tenants (name, slug, settings)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, slug, is_active, settings, created_at`,
		name, slug, settings,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	return &tenant, nil
}

// FindBySlug returns a tenant by its slug.
func (r *TenantRepository) FindBySlug(slug string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Get(&tenant,
		`SELECT id, name, slug, is_active, settings, created_at
		 FROM tenants WHERE slug = $1`, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return &tenant, nil
}

// FindByID returns a tenant by ID.
func (r *TenantRepository) FindByID(id uuid.UUID) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.Get(&tenant,
		`SELECT id, name, slug, is_active, settings, created_at
		 FROM tenants WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return &tenant, nil
}

// UpdateSettings updates tenant settings.
func (r *TenantRepository) UpdateSettings(id uuid.UUID, settings model.TenantSettings) error {
	_, err := r.db.Exec(
		`UPDATE tenants SET settings = $1 WHERE id = $2`,
		settings, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}
	return nil
}

// List returns all active tenants. TODO: pagination for admin dashboard.
func (r *TenantRepository) List() ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := r.db.Select(&tenants,
		`SELECT id, name, slug, is_active, settings, created_at
		 FROM tenants WHERE is_active = true ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}
