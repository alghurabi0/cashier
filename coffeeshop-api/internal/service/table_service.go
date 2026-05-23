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

// List returns all active tables.
func (s *TableService) List() ([]model.Table, error) {
	return s.tableRepo.List()
}

// GetByToken validates and returns a table by its token.
func (s *TableService) GetByToken(token string) (*model.Table, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	return s.tableRepo.GetByToken(token)
}

// Create creates a new table.
func (s *TableService) Create(req model.CreateTableRequest) (*model.Table, error) {
	if req.Number == "" {
		return nil, &ValidationError{Errors: map[string]string{"number": "is required"}}
	}
	return s.tableRepo.Create(req.Number)
}

// Delete soft-deletes a table.
func (s *TableService) Delete(id uuid.UUID) error {
	return s.tableRepo.Delete(id)
}
