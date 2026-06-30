package service

import (
	"coffeeshop-api/internal/model"
	"coffeeshop-api/internal/repository"
	"fmt"
)

// PublicMenuService serves public (no-auth) menu data scoped by table token.
type PublicMenuService struct {
	tableRepo    *repository.TableRepository
	tenantRepo   *repository.TenantRepository
	categoryRepo *repository.CategoryRepository
	menuItemRepo *repository.MenuItemRepository
}

// NewPublicMenuService creates a new PublicMenuService.
func NewPublicMenuService(
	tableRepo *repository.TableRepository,
	tenantRepo *repository.TenantRepository,
	categoryRepo *repository.CategoryRepository,
	menuItemRepo *repository.MenuItemRepository,
) *PublicMenuService {
	return &PublicMenuService{
		tableRepo:    tableRepo,
		tenantRepo:   tenantRepo,
		categoryRepo: categoryRepo,
		menuItemRepo: menuItemRepo,
	}
}

// GetMenu returns all data a menu site needs in a single call.
// Resolves token → table → tenant, then fetches active categories + menu items.
func (s *PublicMenuService) GetMenu(token string) (*model.PublicMenuResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Resolve token to table
	table, err := s.tableRepo.GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid table token")
	}

	// Get tenant info
	tenant, err := s.tenantRepo.FindByID(table.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	if !tenant.IsActive {
		return nil, fmt.Errorf("this shop is currently inactive")
	}

	// Fetch active categories
	allCategories, err := s.categoryRepo.FindAll(table.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to load categories: %w", err)
	}
	// Filter to active only
	var categories []model.Category
	for _, c := range allCategories {
		if c.IsActive {
			categories = append(categories, c)
		}
	}

	// Fetch active menu items
	allItems, err := s.menuItemRepo.FindAll(table.TenantID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load menu items: %w", err)
	}
	// Filter to active only and strip internal fields
	var items []model.MenuItem
	for _, item := range allItems {
		if item.IsActive {
			items = append(items, item.MenuItem)
		}
	}

	return &model.PublicMenuResponse{
		Tenant: model.PublicTenantInfo{
			Name:          tenant.Name,
			Slug:          tenant.Slug,
			IntroVideoURL: tenant.Settings.IntroVideoURL,
		},
		Table: model.PublicTableInfo{
			Number: table.Number,
		},
		Categories: categories,
		MenuItems:  items,
	}, nil
}
