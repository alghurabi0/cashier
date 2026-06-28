package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AdminService handles platform-level admin operations (super_admin only).
type AdminService struct {
	tenantRepo *repository.TenantRepository
	userRepo   *repository.UserRepository
	deviceRepo *repository.DeviceRepository
	adminRepo  *repository.AdminRepository
}

// NewAdminService creates a new AdminService.
func NewAdminService(
	tenantRepo *repository.TenantRepository,
	userRepo *repository.UserRepository,
	deviceRepo *repository.DeviceRepository,
	adminRepo *repository.AdminRepository,
) *AdminService {
	return &AdminService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
		adminRepo:  adminRepo,
	}
}

// ListTenants returns all tenants with user/device counts.
func (s *AdminService) ListTenants() ([]model.TenantWithCounts, error) {
	return s.adminRepo.ListTenantsWithCounts()
}

// GetTenantDetail returns full detail for a single tenant.
func (s *AdminService) GetTenantDetail(id uuid.UUID) (*model.TenantDetail, error) {
	tenant, err := s.tenantRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.ListByTenant(id)
	if err != nil {
		return nil, err
	}

	devices, err := s.deviceRepo.ListByTenant(id)
	if err != nil {
		return nil, err
	}

	return &model.TenantDetail{
		Tenant:  tenant,
		Users:   users,
		Devices: devices,
	}, nil
}

// UpdateTenant updates a tenant's name, active status, and/or settings.
func (s *AdminService) UpdateTenant(id uuid.UUID, req model.UpdateTenantRequest) (*model.Tenant, error) {
	return s.adminRepo.UpdateTenant(id, req)
}

// ListTenantUsers returns all users for a specific tenant.
func (s *AdminService) ListTenantUsers(tenantID uuid.UUID) ([]model.User, error) {
	return s.userRepo.ListByTenant(tenantID)
}

// CreateTenantUser creates a new user under a specific tenant.
func (s *AdminService) CreateTenantUser(tenantID uuid.UUID, req model.CreateUserRequest) (*model.User, error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if req.Role == "" {
		req.Role = "cashier"
	}
	if req.Role != "admin" && req.Role != "cashier" {
		return nil, fmt.Errorf("role must be 'admin' or 'cashier'")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.Create(tenantID, req.Username, string(hash), req.Role)
}

// ListTenantDevices returns all devices for a specific tenant.
func (s *AdminService) ListTenantDevices(tenantID uuid.UUID) ([]model.Device, error) {
	return s.deviceRepo.ListByTenant(tenantID)
}

// GetPlatformStats returns platform-wide statistics.
func (s *AdminService) GetPlatformStats() (*model.PlatformStats, error) {
	return s.adminRepo.GetPlatformStats()
}
