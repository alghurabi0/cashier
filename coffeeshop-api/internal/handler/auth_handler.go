package handler

import (
	"coffeeshop-api/internal/service"
	"net/http"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register handles POST /api/v1/auth/register
// Creates the first admin user. Disabled after the first user exists.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	result, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, result)
}

// Login handles POST /api/v1/auth/login
// Returns a JWT token on valid credentials.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	result, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	JSON(w, http.StatusOK, result)
}
