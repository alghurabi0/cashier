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

type loginRequest struct {
	Username string `json:"username"` // format: "user@tenant-slug"
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login
// Username must be in the format "username@tenant-slug".
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
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
