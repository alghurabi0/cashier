# خطة مشروع تطبيق الكاشير والمخزون — مقهى

**Coffeeshop POS & Inventory — Project Plan v4**

Stack: **Wails v3** · **Vue 3 + TypeScript** · **Go API** · **PostgreSQL** · **SQLite (local cache)** · **WebSocket** · Arabic-first / RTL

## 1. Project overview

A coffeeshop operations system made of three separate but connected projects:

| Project | Type | Deployed on |
| --- | --- | --- |
| coffeeshop-api | Go REST + WebSocket server | Oracle Cloud VPS |
| coffeeshop-pos | Wails v3 desktop app | Windows machines (×2+) |
| coffeeshop-menu | Web app (future) | Oracle Cloud / static host |

**Core principles:**

- The central PostgreSQL database on Oracle is the single source of truth.
- The desktop app has a local SQLite cache and works fully offline — cashier is never blocked.
- Offline actions are queued and sync automatically when the connection returns.
- Separation of Concerns: Menu items (sales) are strictly separated from Inventory items (raw assets), bridged by a Recipe system.
- Web menu orders arrive at the server, are pushed to the desktop via WebSocket, and the cashier accepts or rejects them.

## 2. How the three projects connect

*(Unchanged from v3)*
Orders flow from the Web Menu to the API, pushed via WebSocket to the POS. Cashier orders flow from the POS local SQLite to the API via a background sync worker.

## 3. Offline-first sync design

### Why UUIDs instead of auto-increment IDs

If the cashier creates an order while offline, it needs a stable identity before the server has seen it. Auto-increment IDs would collide. The solution is **UUIDs generated at creation time**, on whichever side creates the record. The server assigns the sequential `order_number` after sync.

### Offline Stock Deduction (The Recipe Engine)

Because the local SQLite cache holds a full copy of the `inventory_items` and `recipe_ingredients`, stock deduction happens locally first:

1. Cashier sells 1x Latte offline.
2. The local Go service looks up the recipe.
3. It deducts 20g of Espresso and 250ml of Milk from the local inventory_items table.
4. The order is queued for sync.
5. Once online, the server processes the order and performs the exact same recipe deduction on the central PostgreSQL database.

### What syncs in which direction

| Data | Direction | Conflict rule |
| --- | --- | --- |
| Menu Items, Categories | Server ↔ Desktop | Last-write wins |
| Inventory Items, Recipes | Server ↔ Desktop | Last-write wins |
| Orders (cash) | Desktop → Server (push) | Local wins |
| Orders (web menu) | Server → Desktop (push/pull) | Server wins |
| Stock adjustments | Desktop → Server | Local wins |
| Settings | Desktop → Server | Last-write wins |

## 4. Shared schema (PostgreSQL — central)

This schema reflects the split between front-of-house (Menu) and back-of-house (Inventory).

SQL```
-- 1. Front-of-House (Cashier)

categories (
  id           UUID PK DEFAULT gen_random_uuid()
  name_ar      TEXT NOT NULL
  sort_order   INT DEFAULT 0
  is_active    BOOL DEFAULT true
)

menu_items (
  id                   UUID PK DEFAULT gen_random_uuid()
  category_id          UUID FK → categories
  name_ar              TEXT NOT NULL
  price                BIGINT NOT NULL       -- fils (IQD × 1000)
  cost_calc_method     TEXT DEFAULT 'auto'   -- 'auto' | 'manual'
  manual_cost_price    BIGINT DEFAULT 0      -- fils
  cached_auto_cost     BIGINT DEFAULT 0      -- updated whenever recipe changes
  image_path           TEXT                  
  is_active            BOOL DEFAULT true
)

-- 2. Back-of-House (Inventory)

inventory_items (
  id                   UUID PK DEFAULT gen_random_uuid()
  name_ar              TEXT NOT NULL         -- e.g., "حليب المراعي", "أكواب ورقية"
  base_unit_ar         TEXT NOT NULL         -- e.g., 'مل', 'غرام', 'قطعة'
  stock_qty            INT DEFAULT 0         -- Always in base unit
  low_stock_threshold  INT DEFAULT 0
  unit_cost            BIGINT DEFAULT 0      -- fils per 1 base unit
  is_active            BOOL DEFAULT true
)

recipe_ingredients (
  id                   UUID PK DEFAULT gen_random_uuid()
  menu_item_id         UUID FK → menu_items ON DELETE CASCADE
  inventory_item_id    UUID FK → inventory_items ON DELETE CASCADE
  quantity             INT NOT NULL          -- Amount of base unit consumed
)

-- 3. Sales & Transactions

orders (
  id             UUID PK DEFAULT gen_random_uuid()
  order_number   TEXT UNIQUE               
  source         TEXT NOT NULL             -- 'cashier' | 'web_menu'
  table_number   TEXT                      
  status         TEXT DEFAULT 'pending'   
  total          BIGINT NOT NULL
  payment_method TEXT                      -- 'cash'|'card'|'web'
  created_at     TIMESTAMPTZ DEFAULT now()
)

order_items (
  id               UUID PK DEFAULT gen_random_uuid()
  order_id         UUID FK → orders ON DELETE CASCADE
  menu_item_id     UUID FK → menu_items    -- nullable if item deleted later
  quantity         INT NOT NULL
  unit_price       BIGINT NOT NULL
  line_total       BIGINT NOT NULL
  name_ar_snapshot TEXT NOT NULL
)

stock_adjustments (
  id                UUID PK DEFAULT gen_random_uuid()
  inventory_item_id UUID FK → inventory_items
  delta             INT NOT NULL           -- positive (receive) or negative (waste)
  reason_ar         TEXT
  created_at        TIMESTAMPTZ DEFAULT now()
)

```

## 5. Central API design (coffeeshop-api)

Endpoints updated to reflect the new architecture.

```
Menu & POS
  GET    /api/menu-items           -- public
  POST   /api/menu-items           -- auth required
  PUT    /api/menu-items/:id

Inventory (Auth Required)
  GET    /api/inventory
  POST   /api/inventory
  PUT    /api/inventory/:id
  POST   /api/inventory/adjust     -- record stock adjustment

Recipes (Auth Required)
  GET    /api/menu-items/:id/recipe
  PUT    /api/menu-items/:id/recipe -- accepts array of ingredients

```

## 6. Project structure & Development workflow

When building out the local SQLite cache and business logic, setting up your environment smoothly is critical.

Plaintext```
coffeeshop/                     ← monorepo root 
  coffeeshop-api/               ← Go API server (Oracle VPS)
  coffeeshop-pos/               ← Wails desktop app
  coffeeshop-menu/              ← Nuxt 3 web app (future)

```

**Development Workflow:**
Opening the `coffeeshop-pos/backend` Go code directly in Antigravity IDE will give you the best structural overview of the sync workers and GORM models. Because Wails supports excellent cross-compilation, you can write and test the entire stack directly on your MacBook Air M3. When you are ready to deploy to the production hardware, running `wails build -platform windows/amd64` from your Mac will compile the clean `.exe` needed for the shop's Windows machines.

## 7. Revised feature phases

### Phase 1 — Central API foundation

- [ ] Go API project scaffold (chi, GORM, PostgreSQL)
- [ ] PostgreSQL schema implementation (separated menu/inventory)
- [ ] Auth endpoints and Caddy reverse proxy on Oracle Cloud
- [ ] Menu and Category endpoints

### Phase 2 — Inventory & Recipe Foundation

- [ ] Inventory items CRUD endpoints
- [ ] Recipe management endpoints
- [ ] Auto-cost calculation logic (Triggered when recipe updates)
- [ ] Desktop app base: Wails v3, SQLite sync skeleton

### Phase 3 — Cashier Core (Front-of-House)

- [ ] Desktop UI: RTL layout, Vue 3 POS grid
- [ ] Cart logic using menu_items
- [ ] Order saved to SQLite (UUID) + local inventory_items decrement logic
- [ ] Background sync worker: push orders to API
- [ ] Receipt printing integration

### Phase 4 — Management & Inventory UI (Back-of-House)

- [ ] Desktop UI: Inventory grid (Raw materials)
- [ ] Recipe builder UI (Link Menu Items to Inventory Items)
- [ ] Stock adjustment flow (adding deliveries, marking waste)
- [ ] Settings panel: toggle Cost Calculation between Auto/Manual

### Phase 5 — Web menu orders

- [ ] WebSocket client in desktop app
- [ ] Incoming orders panel: accept/reject flow
- [ ] Orders API endpoint ready for web menu (table token auth)

### Phase 6 — Reports & Polish

- [ ] Order history, voiding
- [ ] Daily profit reports (Total Sales - Total Recipe Costs)
- [ ] PIN auth for admin vs. cashier roles
- [ ] Windows cross-compilation and deployment

## 8. Key design decisions summary

| Decision | Choice | Reason |
| --- | --- | --- |
| Item Separation | Menu vs. Inventory | accurately handles recipes and direct sales (1-to-1 recipe). |
| Stock Math | Smallest Base Unit | Eliminates floating-point errors and fraction management. |
| Costing | Auto by default | The system does the math (Qty * Unit Cost) but allows manual override. |
| Offline Strategy | Full local simulation | Desktop app knows recipes and deducts raw stock locally for real-time safety. |
| IDs | UUIDs everywhere | Offline record creation without ID collision. |