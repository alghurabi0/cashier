package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TenantService handles tenant operations.
type TenantService struct {
	tenantRepo *repository.TenantRepository
	userRepo   *repository.UserRepository
	authSvc    *AuthService
}

// NewTenantService creates a new TenantService.
func NewTenantService(tenantRepo *repository.TenantRepository, userRepo *repository.UserRepository, authSvc *AuthService) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		authSvc:    authSvc,
	}
}

// Create creates a new tenant with its first admin user, returning the tenant, user, and JWT.
func (s *TenantService) Create(req model.CreateTenantRequest) (*model.CreateTenantResponse, error) {
	// Validate
	if req.Name == "" {
		return nil, fmt.Errorf("tenant name is required")
	}
	if req.Slug == "" {
		return nil, fmt.Errorf("tenant slug is required")
	}
	if req.AdminUsername == "" {
		return nil, fmt.Errorf("admin username is required")
	}
	if len(req.AdminPassword) < 6 {
		return nil, fmt.Errorf("admin password must be at least 6 characters")
	}

	// Create tenant with default settings
	defaultSettings := model.TenantSettings{
		KitchenModeEnabled:     false,
		ConflictResolutionMode: "last-write-wins",
	}
	tenant, err := s.tenantRepo.Create(req.Name, req.Slug, defaultSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create admin user
	hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.Create(tenant.ID, req.AdminUsername, string(hash), "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Generate JWT
	token, err := s.authSvc.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.CreateTenantResponse{
		Tenant: tenant,
		Token:  token,
		User:   user,
	}, nil
}

// GetBySlug returns a tenant by slug (public lookup for POS setup).
func (s *TenantService) GetBySlug(slug string) (*model.Tenant, error) {
	return s.tenantRepo.FindBySlug(slug)
}

// GetSettings returns the settings for a tenant.
func (s *TenantService) GetSettings(tenantID uuid.UUID) (*model.TenantSettings, error) {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return nil, err
	}
	return &tenant.Settings, nil
}

// UpdateSettings updates the settings for a tenant.
func (s *TenantService) UpdateSettings(tenantID uuid.UUID, settings model.TenantSettings) error {
	return s.tenantRepo.UpdateSettings(tenantID, settings)
}
