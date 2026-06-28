package handler

import (
	"coffeeshop-api/internal/service"
	"net/http"
)

// PublicMenuHandler serves public (no-auth) menu data.
type PublicMenuHandler struct {
	publicMenuService *service.PublicMenuService
}

// NewPublicMenuHandler creates a new PublicMenuHandler.
func NewPublicMenuHandler(publicMenuService *service.PublicMenuService) *PublicMenuHandler {
	return &PublicMenuHandler{publicMenuService: publicMenuService}
}

// GetMenu handles GET /api/v1/public/menu?token={table_token}.
// No authentication required — the table token scopes the data.
func (h *PublicMenuHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		Error(w, http.StatusBadRequest, "token query parameter is required")
		return
	}

	menu, err := h.publicMenuService.GetMenu(token)
	if err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, menu)
}
