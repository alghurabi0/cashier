package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/repository"
	"net/http"
)

// DeviceHandler handles device registration and listing.
type DeviceHandler struct {
	deviceRepo *repository.DeviceRepository
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(deviceRepo *repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{deviceRepo: deviceRepo}
}

type registerDeviceRequest struct {
	DeviceName string `json:"device_name"`
	DeviceType string `json:"device_type"` // "pos" (default), "kitchen_display", "web"
}

// Register handles POST /api/v1/devices/register
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req registerDeviceRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	if req.DeviceName == "" {
		Error(w, http.StatusBadRequest, "device_name is required")
		return
	}

	device, err := h.deviceRepo.Register(tenantID, req.DeviceName, req.DeviceType)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, device)
}

// List handles GET /api/v1/devices
func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	devices, err := h.deviceRepo.ListByTenant(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, devices)
}
