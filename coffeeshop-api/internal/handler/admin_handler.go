package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"

	"github.com/google/uuid"
)

// AdminHandler handles platform-level admin endpoints (super_admin only).
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ListTenants handles GET /api/v1/admin/tenants.
func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := h.adminService.ListTenants()
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, tenants)
}

// GetTenant handles GET /api/v1/admin/tenants/{id}.
func (h *AdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	detail, err := h.adminService.GetTenantDetail(id)
	if err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}
	JSON(w, http.StatusOK, detail)
}

// UpdateTenant handles PUT /api/v1/admin/tenants/{id}.
func (h *AdminHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var req model.UpdateTenantRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	tenant, err := h.adminService.UpdateTenant(id, req)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, tenant)
}

// ListTenantUsers handles GET /api/v1/admin/tenants/{id}/users.
func (h *AdminHandler) ListTenantUsers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	users, err := h.adminService.ListTenantUsers(id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, users)
}

// CreateTenantUser handles POST /api/v1/admin/tenants/{id}/users.
func (h *AdminHandler) CreateTenantUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var req model.CreateUserRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	user, err := h.adminService.CreateTenantUser(id, req)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusCreated, user)
}

// ListTenantDevices handles GET /api/v1/admin/tenants/{id}/devices.
func (h *AdminHandler) ListTenantDevices(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	devices, err := h.adminService.ListTenantDevices(id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, devices)
}

// GetStats handles GET /api/v1/admin/stats.
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetUserID(r.Context()) // ensure authenticated

	stats, err := h.adminService.GetPlatformStats()
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, stats)
}
