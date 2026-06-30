package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

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

// GenerateProvisionCode creates a random 6-char alphanumeric code for a tenant.
// The code expires in 24 hours.
func (s *TenantService) GenerateProvisionCode(tenantID uuid.UUID) (string, time.Time, error) {
	code, err := randomCode(6)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate code: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.tenantRepo.SetProvisionCode(tenantID, code, &expiresAt); err != nil {
		return "", time.Time{}, err
	}

	return code, expiresAt, nil
}

// Provision looks up a tenant by provision code and returns auth credentials.
// If the tenant has no users, creates a default admin. Otherwise returns the first admin's credentials.
func (s *TenantService) Provision(code string) (*model.ProvisionResponse, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil, fmt.Errorf("provision code is required")
	}

	tenant, err := s.tenantRepo.FindByProvisionCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired provision code")
	}

	// Generate a random password for the provisioned user
	password, err := randomCode(10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	username := "admin"

	// Check if any users exist for this tenant
	count, err := s.userRepo.Count(tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check users: %w", err)
	}

	var user *model.User
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user, err = s.userRepo.Create(tenant.ID, username, string(hash), "admin")
		if err != nil {
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	} else {
		user, err = s.userRepo.FindByUsername(tenant.ID, username)
		if err != nil {
			return nil, fmt.Errorf("admin user not found: %w", err)
		}
		// Update password for reprovisioning
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		if err := s.userRepo.UpdatePassword(user.ID, string(hash)); err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	token, err := s.authSvc.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Clear the provision code after use (single-use)
	_ = s.tenantRepo.ClearProvisionCode(tenant.ID)

	return &model.ProvisionResponse{
		Tenant:   tenant,
		Token:    token,
		Username: username + "@" + tenant.Slug,
		Password: password,
	}, nil
}

func randomCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
