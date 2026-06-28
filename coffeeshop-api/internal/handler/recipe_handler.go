package handler

import (
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RecipeHandler handles HTTP requests for recipe operations.
type RecipeHandler struct {
	recipeService *service.RecipeService
}

// NewRecipeHandler creates a new RecipeHandler.
func NewRecipeHandler(recipeService *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService}
}

// ListAll handles GET /api/v1/recipes
func (h *RecipeHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())

	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid 'since' parameter: use ISO 8601 format")
			return
		}
		ingredients, err := h.recipeService.GetAllRecipesSince(tenantID, since)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to fetch recipes")
			return
		}
		JSON(w, http.StatusOK, ingredients)
		return
	}

	ingredients, err := h.recipeService.GetAllRecipes(tenantID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch recipes")
		return
	}
	JSON(w, http.StatusOK, ingredients)
}

// Get handles GET /api/v1/menu-items/{id}/recipe
func (h *RecipeHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	menuItemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	ingredients, err := h.recipeService.GetRecipe(tenantID, menuItemID)
	if err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, ingredients)
}

// Set handles PUT /api/v1/menu-items/{id}/recipe
func (h *RecipeHandler) Set(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	menuItemID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid menu item ID")
		return
	}

	var req model.SetRecipeRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	ingredients, err := h.recipeService.SetRecipe(tenantID, menuItemID, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusBadRequest, err.Error())
		return
	}

	JSON(w, http.StatusOK, ingredients)
}
