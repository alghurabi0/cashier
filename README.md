# ☕ Coffeeshop POS — Developer Guide

## Overview

A full-stack coffeeshop point-of-sale system with 3 components:

| Component | Tech | Port | Description |
|---|---|---|---|
| **coffeeshop-api** | Go + PostgreSQL | `:8080` | Central REST API |
| **coffeeshop-pos** | Go + Wails v3 + Vue 3 | Native | Desktop POS app |
| **coffeeshop-menu** | Vue 3 + Vite | `:5173` | Customer-facing web menu |

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ coffeeshop-  │     │ coffeeshop-  │     │ coffeeshop-  │
│    menu      │────▶│     api      │◀────│     pos      │
│  (Customer)  │     │  (Server)    │     │  (Desktop)   │
└──────────────┘     └──────────────┘     └──────────────┘
     :5173                :8080              Wails Native
                           │
                    ┌──────┴──────┐
                    │  PostgreSQL │
                    │    :5432    │
                    └─────────────┘
```

---

## Prerequisites

| Tool | Version | Required For |
|---|---|---|
| [Docker](https://docs.docker.com/get-docker/) | 20+ | All backend services |
| [Go](https://go.dev/dl/) | 1.24+ | POS desktop app |
| [Node.js](https://nodejs.org/) | 20+ | POS frontend |
| [Wails v3](https://v3.wails.io/) | v3.0.0-alpha | POS desktop app |
| [golang-migrate](https://github.com/golang-migrate/migrate) | v4 | Database migrations (CLI) |

### Install Wails v3

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Install golang-migrate (optional — Docker handles this)

```bash
brew install golang-migrate
```

---

## Quick Start (Docker)

The fastest way to get everything running:

```bash
# 1. Clone and navigate to the project
cd cashier

# 2. Start PostgreSQL + Migrations + API + Web Menu
docker compose up -d

# 3. Seed the database with sample data
docker compose exec api go run ./seed/...

# 4. Create an API user (the POS uses this to talk to the API)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username": "admin", "password": "admin123"}'

# 5. Start the POS desktop app (runs natively)
cd coffeeshop-pos
npm install --prefix frontend
wails3 dev
```

After these steps:
- **API** is running at `http://localhost:8080`
- **Web Menu** is running at `http://localhost:5173`
- **POS** is running as a native desktop window
- **POS login PIN**: `0000` (default admin, created on first launch)

> **Two auth layers explained:**
> - **API credentials** → Entered once in the setup wizard (stored locally, connects POS to central API)
> - **PIN `0000`** → Daily login (controls admin vs cashier UI access, works offline)
>
> On first launch, the POS shows a setup wizard. After that, it auto-connects and you only need your PIN.

---

## Services Deep Dive

### 1. PostgreSQL Database

Started automatically by Docker Compose.

| Setting | Value |
|---|---|
| Host | `localhost` |
| Port | `5432` |
| Database | `coffeeshop` |
| User | `coffeeshop` |
| Password | `password` |

**Connect manually:**

```bash
docker compose exec postgres psql -U coffeeshop
```

**Reset database (destructive):**

```bash
docker compose down -v  # removes volumes
docker compose up -d    # recreates everything
```

---

### 2. Database Migrations

Migrations run automatically when `docker compose up` starts. They use [golang-migrate](https://github.com/golang-migrate/migrate).

**Migrations are located in:** `coffeeshop-api/migrations/`

| Migration | Description |
|---|---|
| `000001_create_users` | API authentication users |
| `000002_create_categories` | Menu categories |
| `000003_create_menu_items` | Menu products |
| `000004_create_inventory_items` | Raw materials |
| `000005_create_recipe_ingredients` | Recipes linking menu → inventory |
| `000006_create_orders` | Orders + order items |
| `000007_create_stock_adjustments` | Inventory adjustments log |
| `000008_create_tables` | Restaurant tables (web orders) |

**Run migrations manually (without Docker):**

```bash
export DATABASE_URL="postgres://coffeeshop:password@localhost:5432/coffeeshop?sslmode=disable"

# Apply all pending
cd coffeeshop-api
make migrate-up

# Roll back last migration
make migrate-down

# Create a new migration
make migrate-create name=create_xyz
```

---

### 3. Seed Data

The seed script populates the database with sample coffeeshop data:

- **3 categories**: مشروبات ساخنة, مشروبات باردة, حلويات
- **9 menu items**: اسبريسو, لاتيه, كابتشينو, شاي, آيس لاتيه, etc.
- **6 inventory items**: حبوب قهوة, حليب, سكر, أكواب ورقية, etc.
- **7 recipes**: linking menu items to their raw ingredients with quantities

**Run seed (Docker):**

```bash
docker compose exec api go run ./seed/...
```

**Run seed (native):**

```bash
cd coffeeshop-api
make seed
```

> The seed script is **idempotent** — running it multiple times won't create duplicates.

---

### 4. Coffeeshop API

The central REST API that all clients connect to.

**Run with Docker (recommended):**

```bash
docker compose up api -d
```

**Run natively:**

```bash
cd coffeeshop-api
cp .env.example .env  # Edit as needed
make run
```

**Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `DATABASE_URL` | — | PostgreSQL connection string (required) |
| `JWT_SECRET` | — | Secret for JWT token signing (required) |
| `ENVIRONMENT` | `development` | `development` or `production` |

**API endpoints overview:**

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | — | Create API user |
| POST | `/api/v1/auth/login` | — | Login → JWT token |
| GET | `/api/v1/categories` | — | List categories |
| GET | `/api/v1/menu-items` | — | List menu items |
| GET | `/api/v1/orders` | JWT | Order history (from/to) |
| POST | `/api/v1/orders` | JWT | Sync order from POS |
| PUT | `/api/v1/orders/{id}/status` | JWT | Update order status |
| GET/POST/PUT/DELETE | `/api/v1/inventory` | JWT | Inventory CRUD |
| GET/PUT | `/api/v1/menu-items/{id}/recipe` | JWT | Recipe management |
| GET/POST/DELETE | `/api/v1/tables` | JWT | Table management |
| POST | `/api/v1/web-orders?token=X` | Token | Customer web order |
| GET | `/api/v1/orders/stream` | JWT | SSE event stream |

---

### 5. Coffeeshop POS (Desktop App)

The POS is a **Wails v3** desktop application. It must run natively on the host machine (not in Docker) because it requires a native webview.

**First-time setup:**

```bash
cd coffeeshop-pos

# Install frontend dependencies
npm install --prefix frontend

# Generate Wails bindings (Go → TypeScript)
wails3 generate bindings
```

**Run in development mode:**

```bash
cd coffeeshop-pos
wails3 dev
```

This starts:
- Go backend with hot reload
- Vite dev server (proxied through Wails)
- Auto-regenerates bindings when Go code changes

**Build for production:**

```bash
cd coffeeshop-pos
wails3 build
```

**POS Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `API_BASE_URL` | `http://localhost:8080` | Default API URL (used until setup wizard runs) |
| `DB_PATH` | `coffeeshop.db` | Local SQLite database path |

**First launch — Setup Wizard:**
- The app shows a setup screen to enter API URL + username + password
- Credentials are validated and stored locally in SQLite
- On subsequent launches, credentials are loaded automatically

**Daily login — PIN pad:**
- After setup, users login with their 4-digit PIN
- Default admin PIN is `0000` (auto-created on first launch)
- Change the PIN or add more users in Settings → User Management

**POS Navigation (Admin):**

| Icon | View | Description |
|---|---|---|
| 📋 | نقطة البيع | Point-of-sale (create orders) |
| 🌐 | طلبات الويب | Incoming web orders (accept/reject) |
| 📜 | سجل الطلبات | Order history + voiding |
| 📊 | التقارير | Profit reports |
| 📦 | المخزون | Inventory & recipes management |
| ⚙️ | الإعدادات | Settings, sync, user management |

---

### 6. Web Menu (Customer-Facing)

A mobile-first Vue 3 SPA that customers use to browse and order from their table.

**Run with Docker (recommended):**

```bash
docker compose up menu -d
# → http://localhost:5173
```

**Run natively:**

```bash
cd coffeeshop-menu
npm install
npx vite --host
```

**How customers access it:**

1. Admin creates a table in POS → Settings → (or via API)
2. Each table has a unique token
3. Generate a QR code pointing to: `http://localhost:5173?token=TABLE_TOKEN`
4. Customer scans QR → browses menu → places order
5. Order appears in POS → Web Orders panel

**Environment variables:**

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Central API URL |

---

## Common Development Commands

### Docker Compose

```bash
# Start everything
docker compose up -d

# View logs
docker compose logs -f api
docker compose logs -f menu

# Restart a service
docker compose restart api

# Stop everything
docker compose down

# Stop + delete data
docker compose down -v
```

### Database

```bash
# Connect to PostgreSQL
docker compose exec postgres psql -U coffeeshop

# Run migrations
docker compose run --rm migrate \
  -path=/migrations \
  -database="postgres://coffeeshop:password@postgres:5432/coffeeshop?sslmode=disable" \
  up

# Rollback migration
docker compose run --rm migrate \
  -path=/migrations \
  -database="postgres://coffeeshop:password@postgres:5432/coffeeshop?sslmode=disable" \
  down 1

# Seed data
docker compose exec api go run ./seed/...
```

### API

```bash
# Get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')

# Test an endpoint
curl -s http://localhost:8080/api/v1/categories | jq

# Test auth endpoint
curl -s http://localhost:8080/api/v1/inventory \
  -H "Authorization: Bearer $TOKEN" | jq
```

### POS Desktop

```bash
cd coffeeshop-pos

# Dev mode (hot reload)
wails3 dev

# Regenerate bindings after Go changes
wails3 generate bindings

# Build production binary
wails3 build

# Frontend only (no Go)
cd frontend && npx vite build
```

### Web Menu

```bash
cd coffeeshop-menu

# Dev mode
npx vite --host

# Build for production
npx vite build
```

---

## Project Structure

```
cashier/
├── docker-compose.yml          ← Development orchestration
├── README.md                   ← This file
├── project-plan.md             ← Project planning document
│
├── coffeeshop-api/             ← Central REST API
│   ├── cmd/api/main.go         ← Entry point + route wiring
│   ├── internal/
│   │   ├── config/             ← Environment config
│   │   ├── database/           ← PostgreSQL connection
│   │   ├── handler/            ← HTTP handlers
│   │   ├── middleware/         ← Auth, CORS, logging
│   │   ├── model/              ← Data models
│   │   ├── repository/         ← Database queries
│   │   ├── service/            ← Business logic
│   │   └── sse/                ← Server-Sent Events hub
│   ├── migrations/             ← PostgreSQL migrations (8)
│   ├── seed/                   ← Sample data seeder
│   ├── Makefile                ← Dev commands
│   └── Dockerfile              ← Container build
│
├── coffeeshop-pos/             ← Desktop POS (Wails v3)
│   ├── main.go                 ← Wails app init + service wiring
│   ├── internal/
│   │   ├── config/             ← POS config
│   │   ├── database/           ← SQLite connection
│   │   ├── migration/          ← SQLite schema (local_users, etc.)
│   │   ├── model/              ← POS data models
│   │   ├── service/            ← 7 Wails-bound services
│   │   │   ├── auth_service.go     ← PIN auth + user management
│   │   │   ├── data_service.go     ← CRUD for menu/categories
│   │   │   ├── management_service.go ← Sync + inventory mgmt
│   │   │   ├── order_service.go    ← Order creation + history
│   │   │   ├── receipt_service.go  ← Receipt formatting
│   │   │   ├── report_service.go   ← Profit reports
│   │   │   └── web_order_service.go ← Web order lifecycle
│   │   └── sync/               ← API client + SSE + sync worker
│   └── frontend/               ← Vue 3 + TypeScript
│       ├── src/
│       │   ├── App.vue         ← Root + auth gate
│       │   ├── views/          ← 6 main views
│       │   ├── components/     ← Reusable UI components
│       │   ├── composables/    ← Vue composables (state + API)
│       │   └── types/          ← TypeScript types
│       └── bindings/           ← Auto-generated Go bindings
│
└── coffeeshop-menu/            ← Customer web menu
    ├── src/
    │   ├── App.vue             ← State machine routing
    │   ├── views/              ← Menu, Cart, Confirmation
    │   ├── components/         ← MenuItemCard, CategoryBar
    │   ├── composables/        ← useMenu, useCart
    │   └── types/              ← TypeScript types
    └── Dockerfile.dev          ← Dev container
```

---

## Troubleshooting

### "DATABASE_URL is required"

The API needs a PostgreSQL connection. Ensure Docker is running:

```bash
docker compose up postgres -d
```

### POS can't connect to API

Check that the API is running at `http://localhost:8080`:

```bash
curl http://localhost:8080/api/v1/categories
```

If using Docker, ensure port 8080 is mapped.

### Wails build fails

Make sure Wails v3 is installed and CGO is enabled:

```bash
wails3 doctor
CGO_ENABLED=1 go build ./...
```

### "Module not found" in POS frontend

Regenerate the Wails bindings:

```bash
cd coffeeshop-pos
wails3 generate bindings
```

### Web Menu shows "رابط غير صالح" (Invalid link)

The web menu requires a table token in the URL. Create a table first:

```bash
TOKEN="your-jwt-token"
curl -X POST http://localhost:8080/api/v1/tables \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"number": "1"}'
```

Then use the returned token: `http://localhost:5173?token=TABLE_TOKEN`











POS frontend setup wizard UI update (tenant-slug in login form)
Device registration UI on first POS launch
Admin web dashboard for tenant management (marked as TODO)