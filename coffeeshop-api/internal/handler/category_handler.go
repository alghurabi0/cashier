package handler

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/service"
	"net/http"

	"github.com/google/uuid"
)

// CategoryHandler handles HTTP requests for category operations.
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List handles GET /api/v1/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.List()
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}

	JSON(w, http.StatusOK, categories)
}

// Get handles GET /api/v1/categories/{id}
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	category, err := h.categoryService.Get(id)
	if err != nil {
		Error(w, http.StatusNotFound, "category not found")
		return
	}

	JSON(w, http.StatusOK, category)
}

// Create handles POST /api/v1/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	category, err := h.categoryService.Create(req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusCreated, category)
}

// Update handles PUT /api/v1/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	var req model.UpdateCategoryRequest
	if !DecodeJSON(w, r, &req) {
		return
	}

	category, err := h.categoryService.Update(id, req)
	if err != nil {
		if ve, ok := err.(*service.ValidationError); ok {
			ValidationError(w, ve.Errors)
			return
		}
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	JSON(w, http.StatusOK, category)
}

// Delete handles DELETE /api/v1/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	if err := h.categoryService.Delete(id); err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
