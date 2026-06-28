package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"

	"github.com/google/uuid"
)

// TableHandler handles HTTP requests for table management.
type TableHandler struct {
	tableService *service.TableService
}

// NewTableHandler creates a new TableHandler.
func NewTableHandler(tableService *service.TableService) *TableHandler {
	return &TableHandler{tableService: tableService}
}

// List handles GET /api/v1/tables
func (h *TableHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	tables, err := h.tableService.List(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, tables)
}

// Create handles POST /api/v1/tables
func (h *TableHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	var req model.CreateTableRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	table, err := h.tableService.Create(tenantID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusCreated, table)
}

// Delete handles DELETE /api/v1/tables/{id}
func (h *TableHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid table ID")
		return
	}

	if err := h.tableService.Delete(tenantID, id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
