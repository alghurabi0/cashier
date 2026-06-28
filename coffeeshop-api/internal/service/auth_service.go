package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication and token management.
type AuthService struct {
	userRepo   *repository.UserRepository
	tenantRepo *repository.TenantRepository
	jwtSecret  string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, tenantRepo *repository.TenantRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
		jwtSecret:  jwtSecret,
	}
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token  string       `json:"token"`
	User   *model.User  `json:"user"`
	Tenant *model.Tenant `json:"tenant"`
}

// Register creates the first admin user for a tenant.
// Only works if the tenant has no users yet.
func (s *AuthService) Register(tenantID uuid.UUID, username, password string) (*AuthResponse, error) {
	// Check if any users exist for this tenant
	count, err := s.userRepo.Count(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("registration is disabled: an admin user already exists for this tenant")
	}

	// Validate input
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the admin user
	user, err := s.userRepo.Create(tenantID, username, string(hash), "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Get tenant for response
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user, Tenant: tenant}, nil
}

// Login validates credentials using the user@tenant-slug format.
// Parses the username into (user, slug), resolves the tenant, then authenticates.
func (s *AuthService) Login(fullUsername, password string) (*AuthResponse, error) {
	// Parse "username@tenant-slug" format
	username, tenantSlug, err := parseLogin(fullUsername)
	if err != nil {
		return nil, err
	}

	// Resolve tenant
	tenant, err := s.tenantRepo.FindBySlug(tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if !tenant.IsActive {
		return nil, fmt.Errorf("tenant is deactivated")
	}

	// Find user in tenant
	user, err := s.userRepo.FindByUsername(tenant.ID, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user, Tenant: tenant}, nil
}

// parseLogin splits "username@tenant-slug" into components.
func parseLogin(fullUsername string) (username, tenantSlug string, err error) {
	parts := strings.SplitN(fullUsername, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid username format: use username@tenant-slug")
	}
	return parts[0], parts[1], nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),
		"role":      user.Role,
		"tenant_id": user.TenantID.String(),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
