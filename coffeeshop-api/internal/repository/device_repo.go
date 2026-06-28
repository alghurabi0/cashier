package repository

import (
	"coffeeshop-api/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// DeviceRepository handles database operations for devices.
type DeviceRepository struct {
	db *sqlx.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Register creates a new device under a tenant.
func (r *DeviceRepository) Register(tenantID uuid.UUID, name, deviceType string) (*model.Device, error) {
	if deviceType == "" {
		deviceType = "pos"
	}
	var device model.Device
	err := r.db.Get(&device,
		`INSERT INTO devices (tenant_id, device_name, device_type)
		 VALUES ($1, $2, $3)
		 RETURNING id, tenant_id, device_name, device_type, is_active, last_seen_at, created_at`,
		tenantID, name, deviceType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}
	return &device, nil
}

// FindByID returns a device by ID.
func (r *DeviceRepository) FindByID(id uuid.UUID) (*model.Device, error) {
	var device model.Device
	err := r.db.Get(&device,
		`SELECT id, tenant_id, device_name, device_type, is_active, last_seen_at, created_at
		 FROM devices WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

// TouchLastSeen updates the last_seen_at timestamp.
func (r *DeviceRepository) TouchLastSeen(id uuid.UUID) error {
	_, err := r.db.Exec(
		`UPDATE devices SET last_seen_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}

// ListByTenant returns all active devices for a tenant.
func (r *DeviceRepository) ListByTenant(tenantID uuid.UUID) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Select(&devices,
		`SELECT id, tenant_id, device_name, device_type, is_active, last_seen_at, created_at
		 FROM devices WHERE tenant_id = $1 AND is_active = true ORDER BY device_name ASC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	if devices == nil {
		devices = []model.Device{}
	}
	return devices, nil
}
