.PHONY: run build test seed migrate-up migrate-down migrate-create

run: ## Start the API server
	go run ./cmd/api/...

build: ## Build the API binary
	go build -o bin/coffeeshop-api ./cmd/api/...

test: ## Run all tests
	go test ./...

seed: ## Seed the database with sample data
	go run ./seed/...

migrate-up: ## Run all pending migrations
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down: ## Roll back the last migration
	migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=xxx)
	migrate create -ext sql -dir migrations -seq $(name)
