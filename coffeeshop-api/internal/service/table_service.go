package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

// TableService handles business logic for table management.
type TableService struct {
	tableRepo *repository.TableRepository
}

// NewTableService creates a new TableService.
func NewTableService(tableRepo *repository.TableRepository) *TableService {
	return &TableService{tableRepo: tableRepo}
}

// List returns all active tables for a tenant.
func (s *TableService) List(tenantID uuid.UUID) ([]model.Table, error) {
	return s.tableRepo.List(tenantID)
}

// GetByToken validates and returns a table by its token (no tenant needed — token is globally unique).
func (s *TableService) GetByToken(token string) (*model.Table, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	return s.tableRepo.GetByToken(token)
}

// Create creates a new table under a tenant.
func (s *TableService) Create(tenantID uuid.UUID, req model.CreateTableRequest) (*model.Table, error) {
	if req.Number == "" {
		return nil, &ValidationError{Errors: map[string]string{"number": "is required"}}
	}
	return s.tableRepo.Create(tenantID, req.Number, req.ID)
}

// Delete soft-deletes a table (tenant-scoped).
func (s *TableService) Delete(tenantID uuid.UUID, id uuid.UUID) error {
	return s.tableRepo.Delete(tenantID, id)
}
