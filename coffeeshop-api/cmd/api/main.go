package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coffeeshop-api/internal/config"
	"coffeeshop-api/internal/database"
	"coffeeshop-api/internal/handler"
	"coffeeshop-api/internal/middleware"
	"coffeeshop-api/internal/repository"
	"coffeeshop-api/internal/service"
	"coffeeshop-api/internal/sse"
	"coffeeshop-api/internal/storage"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup structured logging
	logLevel := slog.LevelInfo
	if cfg.IsDevelopment() {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	// Wire dependencies: repositories
	tenantRepo := repository.NewTenantRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	menuItemRepo := repository.NewMenuItemRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	recipeRepo := repository.NewRecipeRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	tableRepo := repository.NewTableRepository(db)

	// Admin repository (cross-tenant queries)
	adminRepo := repository.NewAdminRepository(db)

	// Wire dependencies: services
	authService := service.NewAuthService(userRepo, tenantRepo, cfg.JWTSecret)
	tenantService := service.NewTenantService(tenantRepo, userRepo, authService)
	categoryService := service.NewCategoryService(categoryRepo)
	menuItemService := service.NewMenuItemService(menuItemRepo, categoryRepo)
	inventoryService := service.NewInventoryService(inventoryRepo, recipeRepo, menuItemRepo)
	recipeService := service.NewRecipeService(recipeRepo, menuItemRepo, inventoryRepo)
	orderService := service.NewOrderService(orderRepo)
	tableService := service.NewTableService(tableRepo)
	adminService := service.NewAdminService(tenantRepo, userRepo, deviceRepo, adminRepo)

	// SSE hub (shared between handlers for real-time order events)
	sseHub := sse.NewHub()

	// Wire dependencies: handlers
	authHandler := handler.NewAuthHandler(authService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	categoryHandler := handler.NewCategoryHandler(categoryService, sseHub)
	menuItemHandler := handler.NewMenuItemHandler(menuItemService, sseHub)
	inventoryHandler := handler.NewInventoryHandler(inventoryService, sseHub)
	recipeHandler := handler.NewRecipeHandler(recipeService)
	orderHandler := handler.NewOrderHandler(orderService, sseHub)
	tableHandler := handler.NewTableHandler(tableService)
	webOrderHandler := handler.NewWebOrderHandler(orderService, tableService, sseHub)
	sseHandler := handler.NewSSEHandler(sseHub)
	healthHandler := handler.NewHealthHandler()

	// Device handler (simple registration)
	deviceHandler := handler.NewDeviceHandler(deviceRepo)

	// Admin handler (platform-level, super_admin only)
	adminHandler := handler.NewAdminHandler(adminService)

	// Public menu service + handler (no auth, token-scoped)
	publicMenuService := service.NewPublicMenuService(tableRepo, tenantRepo, categoryRepo, menuItemRepo)
	publicMenuHandler := handler.NewPublicMenuHandler(publicMenuService)

	// R2 storage (nil if not configured)
	r2, err := storage.NewR2Storage()
	if err != nil {
		slog.Warn("R2 storage init failed", "error", err)
	}
	if r2 != nil {
		slog.Info("R2 storage configured")
	} else {
		slog.Warn("R2 storage not configured (uploads will be unavailable)")
	}
	uploadHandler := handler.NewUploadHandler(r2)

	// Auth middleware
	authMw := middleware.Auth(cfg.JWTSecret)

	// Build router
	mux := http.NewServeMux()

	// ─── Public routes (no auth) ───
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("GET /api/v1/health", healthHandler.Check)

	// Tenant self-service signup (creates tenant + admin user + JWT)
	mux.HandleFunc("POST /api/v1/tenants", tenantHandler.Create)
	mux.HandleFunc("GET /api/v1/tenants/{slug}", tenantHandler.GetBySlug)

	// Web orders (table-token auth via query param)
	mux.HandleFunc("POST /api/v1/web-orders", webOrderHandler.Create)

	// Public menu data (table-token scoped, no auth)
	mux.HandleFunc("GET /api/v1/public/menu", publicMenuHandler.GetMenu)

	// ─── Protected routes (JWT auth — all tenant-scoped) ───

	// Tenant settings
	mux.HandleFunc("GET /api/v1/tenant/settings", authMw(tenantHandler.GetSettings))
	mux.HandleFunc("PUT /api/v1/tenant/settings", authMw(tenantHandler.UpdateSettings))

	// Device registration
	mux.HandleFunc("POST /api/v1/devices/register", authMw(deviceHandler.Register))
	mux.HandleFunc("GET /api/v1/devices", authMw(deviceHandler.List))

	// Categories (read routes also require auth now — tenant-scoped data)
	mux.HandleFunc("GET /api/v1/categories", authMw(categoryHandler.List))
	mux.HandleFunc("GET /api/v1/categories/{id}", authMw(categoryHandler.Get))
	mux.HandleFunc("POST /api/v1/categories", authMw(categoryHandler.Create))
	mux.HandleFunc("PUT /api/v1/categories/{id}", authMw(categoryHandler.Update))
	mux.HandleFunc("DELETE /api/v1/categories/{id}", authMw(categoryHandler.Delete))

	// Menu items (read routes require auth — tenant-scoped)
	mux.HandleFunc("GET /api/v1/menu-items", authMw(menuItemHandler.List))
	mux.HandleFunc("GET /api/v1/menu-items/{id}", authMw(menuItemHandler.Get))
	mux.HandleFunc("POST /api/v1/menu-items", authMw(menuItemHandler.Create))
	mux.HandleFunc("PUT /api/v1/menu-items/{id}", authMw(menuItemHandler.Update))
	mux.HandleFunc("DELETE /api/v1/menu-items/{id}", authMw(menuItemHandler.Delete))

	// Inventory
	mux.HandleFunc("GET /api/v1/inventory", authMw(inventoryHandler.List))
	mux.HandleFunc("GET /api/v1/inventory/{id}", authMw(inventoryHandler.Get))
	mux.HandleFunc("POST /api/v1/inventory", authMw(inventoryHandler.Create))
	mux.HandleFunc("PUT /api/v1/inventory/{id}", authMw(inventoryHandler.Update))
	mux.HandleFunc("DELETE /api/v1/inventory/{id}", authMw(inventoryHandler.Delete))
	mux.HandleFunc("POST /api/v1/inventory/adjust", authMw(inventoryHandler.Adjust))

	// Orders
	mux.HandleFunc("GET /api/v1/orders", authMw(orderHandler.List))
	mux.HandleFunc("POST /api/v1/orders", authMw(orderHandler.Create))
	mux.HandleFunc("PUT /api/v1/orders/{id}/status", authMw(orderHandler.UpdateStatus))

	// Recipes
	mux.HandleFunc("GET /api/v1/recipes", authMw(recipeHandler.ListAll))
	mux.HandleFunc("GET /api/v1/menu-items/{id}/recipe", authMw(recipeHandler.Get))
	mux.HandleFunc("PUT /api/v1/menu-items/{id}/recipe", authMw(recipeHandler.Set))

	// Tables
	mux.HandleFunc("GET /api/v1/tables", authMw(tableHandler.List))
	mux.HandleFunc("POST /api/v1/tables", authMw(tableHandler.Create))
	mux.HandleFunc("DELETE /api/v1/tables/{id}", authMw(tableHandler.Delete))

	// SSE stream
	mux.HandleFunc("GET /api/v1/orders/stream", authMw(sseHandler.Stream))

	// File uploads
	mux.HandleFunc("POST /api/v1/uploads", authMw(uploadHandler.Upload))

	// Image proxy (public — keys are UUIDs, effectively unguessable)
	imageHandler := handler.NewImageHandler(r2)
	mux.HandleFunc("GET /api/v1/images/{path...}", imageHandler.Serve)

	// ─── Admin routes (super_admin only) ───
	adminMw := func(h http.HandlerFunc) http.HandlerFunc {
		return authMw(middleware.AdminOnly(h))
	}
	mux.HandleFunc("GET /api/v1/admin/tenants", adminMw(adminHandler.ListTenants))
	mux.HandleFunc("GET /api/v1/admin/tenants/{id}", adminMw(adminHandler.GetTenant))
	mux.HandleFunc("PUT /api/v1/admin/tenants/{id}", adminMw(adminHandler.UpdateTenant))
	mux.HandleFunc("GET /api/v1/admin/tenants/{id}/users", adminMw(adminHandler.ListTenantUsers))
	mux.HandleFunc("POST /api/v1/admin/tenants/{id}/users", adminMw(adminHandler.CreateTenantUser))
	mux.HandleFunc("GET /api/v1/admin/tenants/{id}/devices", adminMw(adminHandler.ListTenantDevices))
	mux.HandleFunc("GET /api/v1/admin/stats", adminMw(adminHandler.GetStats))

	// Apply global middleware
	finalHandler := middleware.Chain(mux,
		middleware.Logger,
		middleware.Recoverer,
		middleware.CORS,
		middleware.Gzip,
	)

	// Create server
	srv := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     finalHandler,
		IdleTimeout: 60 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("server starting", "port", cfg.Port, "env", cfg.Environment)
		fmt.Printf("\n  🚀 Coffeeshop API running at http://localhost:%s\n\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
