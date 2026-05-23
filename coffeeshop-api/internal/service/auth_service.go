package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication and token management.
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// AuthResponse is returned after successful login or registration.
type AuthResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Register creates a new admin user. Only works if no users exist yet.
func (s *AuthService) Register(username, password string) (*AuthResponse, error) {
	// Check if any users exist already
	count, err := s.userRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("registration is disabled: an admin user already exists")
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
	user, err := s.userRepo.Create(username, string(hash), "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// Login validates credentials and returns a JWT token.
func (s *AuthService) Login(username, password string) (*AuthResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
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

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
