package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"
)

// TenantHandler handles tenant-related HTTP endpoints.
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// Create handles POST /api/v1/tenants (unauthenticated — self-service signup).
// Creates a new tenant with its first admin user.
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTenantRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	result, err := h.tenantService.Create(req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, result)
}

// GetBySlug handles GET /api/v1/tenants/{slug} (unauthenticated — POS setup).
func (h *TenantHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		Error(w, http.StatusBadRequest, "slug is required")
		return
	}

	tenant, err := h.tenantService.GetBySlug(slug)
	if err != nil {
		Error(w, http.StatusNotFound, "tenant not found")
		return
	}

	JSON(w, http.StatusOK, tenant)
}

// GetSettings handles GET /api/v1/tenant/settings (authenticated — returns current tenant's settings).
func (h *TenantHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	settings, err := h.tenantService.GetSettings(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, settings)
}

// UpdateSettings handles PUT /api/v1/tenant/settings (authenticated).
func (h *TenantHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var settings model.TenantSettings
	if !DecodeJSON(w, r, &settings) {
		return
	}

	if err := h.tenantService.UpdateSettings(tenantID, settings); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, settings)
}
