package model

import "github.com/google/uuid"

// Category represents a front-of-house menu category (e.g. "مشروبات ساخنة").
type Category struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	NameAr    string    `db:"name_ar"    json:"name_ar"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
}

// CreateCategoryRequest is the expected JSON body for creating a category.
type CreateCategoryRequest struct {
	NameAr    string `json:"name_ar"`
	SortOrder int    `json:"sort_order"`
}

// UpdateCategoryRequest is the expected JSON body for updating a category.
type UpdateCategoryRequest struct {
	NameAr    *string `json:"name_ar,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}
