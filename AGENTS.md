# AGENTS.md — Wails v3 Framework Reference

> **Purpose**: This document is a comprehensive reference for AI coding agents working on this project. Wails v3 is a relatively new framework and most AI models have limited training data for it. This file contains the critical patterns, APIs, and conventions needed to produce correct Wails v3 code.
>
> **Official Docs**: https://v3.wails.io/
>
> **Status**: Wails v3 is currently in **Alpha** (v3.0.0-alpha.x). APIs may evolve.

---

## Table of Contents

1. [What is Wails v3](#what-is-wails-v3)
2. [Architecture Overview](#architecture-overview)
3. [Project Structure](#project-structure)
4. [Go Backend — Core Concepts](#go-backend--core-concepts)
5. [Services (The Binding System)](#services-the-binding-system)
6. [Application Lifecycle](#application-lifecycle)
7. [Binding Generation](#binding-generation)
8. [Frontend — Runtime API](#frontend--runtime-api)
9. [Events System](#events-system)
10. [Window Management](#window-management)
11. [Dialogs](#dialogs)
12. [Menus](#menus)
13. [System Tray](#system-tray)
14. [Drag & Drop and Frameless Windows](#drag--drop-and-frameless-windows)
15. [Asset Embedding](#asset-embedding)
16. [CLI Commands](#cli-commands)
17. [Build System (Taskfile)](#build-system-taskfile)
18. [Platform-Specific Configuration](#platform-specific-configuration)
19. [Common Patterns & Best Practices](#common-patterns--best-practices)
20. [Critical Mistakes to Avoid](#critical-mistakes-to-avoid)
21. [This Project's Specifics](#this-projects-specifics)

---

## What is Wails v3

Wails v3 is a Go framework for building cross-platform desktop applications. It uses a **Go backend** + **Web frontend** architecture. The frontend runs inside a native webview (not Electron — no bundled Chromium). The Go backend and frontend communicate through:

- **Bound services** (Go methods callable from JS/TS as async functions)
- **Events** (bidirectional pub/sub between Go and JS)

Key differences from Electron:
- No Node.js runtime — the backend is pure Go
- Much smaller binary sizes (~8MB vs ~150MB for Electron)
- Native OS webview (WebKit on macOS, WebView2 on Windows, WebKitGTK on Linux)

---

## Architecture Overview

```
┌──────────────────────────────────────────────────┐
│                  Wails Application               │
│                                                  │
│  ┌─────────────────────┐  ┌───────────────────┐  │
│  │     Go Backend       │  │    Web Frontend   │  │
│  │                      │  │                   │  │
│  │  ┌───────────────┐   │  │  Vue/React/Svelte │  │
│  │  │  Services      │◄─┼──┤  or Vanilla JS    │  │
│  │  │  (bound methods)│──┼─►│                   │  │
│  │  └───────────────┘   │  │  @wailsio/runtime │  │
│  │                      │  │                   │  │
│  │  ┌───────────────┐   │  │  Generated        │  │
│  │  │  Events        │◄─┼──┤  Bindings (TS)    │  │
│  │  │  (pub/sub)     │──┼─►│                   │  │
│  │  └───────────────┘   │  └───────────────────┘  │
│  │                      │          │               │
│  │  ┌───────────────┐   │    ┌─────┴─────┐        │
│  │  │  App Lifecycle │   │    │  Webview   │        │
│  │  │  Windows       │   │    │  (Native)  │        │
│  │  │  Menus/Dialogs │   │    └───────────┘        │
│  │  └───────────────┘   │                          │
│  └─────────────────────┘                           │
└──────────────────────────────────────────────────┘
```

**Communication flow:**
1. Frontend calls a generated binding function (e.g., `GreetService.Greet("Alice")`)
2. The Wails bridge serializes the call as JSON and sends it to Go
3. Go executes the method and returns the result through the bridge
4. The JS Promise resolves with the typed result

This is an **in-memory bridge** — no HTTP overhead.

---

## Project Structure

```
cashier-v3/
├── AGENTS.md                 # This file
├── main.go                   # Go entry point — app init, window creation, app.Run()
├── greetservice.go           # Example service (bound to frontend)
├── go.mod / go.sum           # Go module (github.com/wailsapp/wails/v3)
├── Taskfile.yml              # Build system entry point (uses Task runner)
├── build/                    # Build configs per platform
│   ├── config.yml            # Wails build config
│   ├── appicon.png           # App icon
│   ├── Taskfile.yml          # Common build tasks
│   ├── darwin/               # macOS-specific build tasks & configs
│   ├── windows/              # Windows-specific (NSIS, etc.)
│   └── linux/                # Linux-specific (AppImage, deb, etc.)
├── frontend/                 # Web frontend (Vue + Vite + TypeScript)
│   ├── index.html            # HTML entry point
│   ├── package.json          # NPM deps (includes @wailsio/runtime)
│   ├── vite.config.ts        # Vite config with Wails plugin
│   ├── tsconfig.json         # TypeScript config
│   ├── src/                  # Frontend source code
│   │   ├── main.ts           # Frontend entry point
│   │   ├── App.vue           # Root Vue component
│   │   └── components/       # Vue components
│   ├── bindings/             # AUTO-GENERATED — do not edit manually
│   │   ├── changeme/         # Module bindings (matches Go module name)
│   │   └── github.com/       # External package bindings
│   ├── dist/                 # Built frontend (embedded into Go binary)
│   └── public/               # Static assets
└── bin/                      # Compiled binaries output
```

> **IMPORTANT**: The `frontend/bindings/` directory is **auto-generated** by `wails3 generate bindings`. Never edit these files manually. They are regenerated on every build or when you run the generator.

---

## Go Backend — Core Concepts

### Import Path

```go
import "github.com/wailsapp/wails/v3/pkg/application"
```

This is the **only** import you need for the core Wails API. The version in `go.mod` should be `v3.0.0-alpha.x`.

### Application Initialization (Procedural API)

Wails v3 uses a **procedural** API (not declarative like v2). You explicitly:
1. Create the application
2. Create windows
3. Run the application

```go
package main

import (
    "embed"
    "log"

    "github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := application.New(application.Options{
        Name:        "My App",
        Description: "My Wails v3 Application",
        Services: []application.Service{
            application.NewService(&MyService{}),
        },
        Assets: application.AssetOptions{
            Handler: application.AssetFileServerFS(assets),
        },
        Mac: application.MacOptions{
            ApplicationShouldTerminateAfterLastWindowClosed: true,
        },
    })

    app.Window.NewWithOptions(application.WebviewWindowOptions{
        Title:  "Main Window",
        Width:  1024,
        Height: 768,
        URL:    "/",
    })

    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Services (The Binding System)

Services are Go structs whose **exported methods** (uppercase first letter) become callable from the frontend.

### Defining a Service

```go
// services/user_service.go
package services

import "context"

type UserService struct {
    db *Database // inject dependencies via constructor
}

func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}

// Exported methods become frontend-callable bindings
func (s *UserService) GetUser(id int) (*User, error) {
    return s.db.FindUser(id)
}

// Unexported methods (lowercase) are NOT exposed to frontend
func (s *UserService) validateUser(u *User) error {
    // internal only
}
```

### Registering Services

```go
app := application.New(application.Options{
    Services: []application.Service{
        application.NewService(NewUserService(db)),
        application.NewService(&AnotherService{}),
    },
})
```

### Service Lifecycle Methods (Optional)

Services can implement lifecycle interfaces:

```go
// Called when the app starts — services start in REGISTRATION ORDER
func (s *UserService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
    // Initialize resources, run migrations, connect to DB
    // The ctx is cancelled when the app shuts down
    // Returning an error here PREVENTS the app from starting
    return nil
}

// Called when the app shuts down — services shut down in REVERSE registration order
func (s *UserService) ServiceShutdown() error {
    // Close DB connections, flush buffers, save state
    // The app context is ALREADY cancelled at this point
    return nil
}
```

### Error Handling in Services

Always return `error` as the **last return value**. Wails automatically converts Go errors into rejected JS Promises:

```go
// Go
func (s *UserService) GetUser(id int) (*User, error) {
    user, err := s.db.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }
    return user, nil
}
```

```typescript
// Frontend — the error becomes a rejected promise
try {
    const user = await UserService.GetUser(123);
} catch (err) {
    console.error(err); // "user not found: ..."
}
```

---

## Application Lifecycle

```
App Created → Services Start (in order) → Windows Created → App Running
                                                                ↓
App Shutdown ← Services Shutdown (reverse order) ← Windows Closed
```

- **ServiceStartup**: Called in registration order. If any returns an error, the app fails to start.
- **ServiceShutdown**: Called in reverse registration order. Errors are logged but don't prevent shutdown.
- The `context.Context` passed to `ServiceStartup` is cancelled when the app begins shutting down.

---

## Binding Generation

### How It Works

Wails v3 uses a **static analyzer** (not runtime reflection) to examine your Go code and generate type-safe JS/TS bindings.

### Generate Command

```bash
wails3 generate bindings
```

### What Gets Generated

For each registered service, the generator creates:
- **Method bindings**: JS/TS functions that call the corresponding Go methods
- **Model types**: TypeScript interfaces/types mirroring Go structs used in method signatures

### Output Location

Bindings are generated into `frontend/bindings/` organized by Go module path:

```
frontend/bindings/
├── changeme/                    # Your Go module's bindings
│   └── greetservice.ts          # Generated from GreetService
└── github.com/
    └── wailsapp/wails/v3/...   # Wails internal bindings
```

### Calling Bindings from Frontend

```typescript
// Import the generated binding
import { Greet } from '../bindings/changeme/GreetService';

// Call it like a regular async function
const message = await Greet("World");
console.log(message); // "Hello World!"
```

### Auto-Regeneration in Dev Mode

When running `wails3 dev`, bindings are **automatically regenerated** whenever you change your Go code. No manual re-run needed.

### Configuration Flags

```bash
wails3 generate bindings -ts        # Force TypeScript output
wails3 generate bindings -d ./out   # Custom output directory
```

---

## Frontend — Runtime API

### Installation

The runtime is already installed in this project:

```json
{
    "dependencies": {
        "@wailsio/runtime": "latest"
    }
}
```

### Available Modules

```typescript
import { Events, Window, Clipboard, Browser, Screens } from '@wailsio/runtime';
```

### Side-Effect Import (Important!)

Even if you don't use the API directly, include this import to ensure the runtime initializes correctly (needed for window dragging, context menus, etc.):

```typescript
import '@wailsio/runtime';
```

### Window API (from frontend)

```typescript
import { Window } from '@wailsio/runtime';

// Get current window
const win = Window.Get('');

// Control window
win.SetTitle("New Title");
win.Minimise();
win.Maximise();
win.Fullscreen();
win.Centre();
win.SetSize(800, 600);
win.SetPosition(100, 100);
win.Show();
win.Hide();
win.Close();
```

### Clipboard API

```typescript
import { Clipboard } from '@wailsio/runtime';

const text = await Clipboard.GetText();
await Clipboard.SetText("Hello clipboard");
```

### Browser API

```typescript
import { Browser } from '@wailsio/runtime';

Browser.OpenURL("https://example.com"); // Opens in system browser
```

### Screens API

```typescript
import { Screens } from '@wailsio/runtime';

const screens = await Screens.GetAll();
const primary = await Screens.GetPrimary();
```

---

## Events System

Wails v3 provides a unified, bidirectional event system shared between Go and JavaScript.

### Registering Typed Events (Go — Optional but Recommended)

```go
func init() {
    // Register a custom event with its data type
    // This provides type-safe bindings in the frontend
    application.RegisterEvent[string]("time")
    application.RegisterEvent[OrderData]("order-created")
}
```

### Emitting Events

**From Go:**
```go
app.Event.Emit("user-logged-in", map[string]interface{}{
    "userId": 123,
    "name":   "Alice",
})
```

**From Frontend:**
```typescript
import { Events } from '@wailsio/runtime';

Events.Emit("button-clicked", { buttonId: "submit" });
```

### Listening for Events

**From Go:**
```go
app.Event.On("order-created", func(e *application.CustomEvent) {
    order := e.Data.(*Order)
    // Handle event
})
```

**From Frontend:**
```typescript
import { Events } from '@wailsio/runtime';

Events.On("user-logged-in", (event) => {
    console.log(`User ${event.data.name} logged in`);
});
```

### Event Namespaces

- `common:` — Cross-platform events (e.g., `common:WindowFocus`)
- `windows:` — Windows-specific events
- `mac:` — macOS-specific events
- `linux:` — Linux-specific events

### Hooks vs Listeners

- **Listeners** (`On`): Asynchronous event handlers
- **Hooks** (`RegisterHook`): Synchronous handlers that can **intercept and cancel** system events (e.g., `WindowClosing`)

---

## Window Management

### Creating Windows

```go
// Simple window
win := app.Window.New()

// Window with options
win := app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title:            "My Window",
    Width:            1024,
    Height:           768,
    URL:              "/",
    BackgroundColour: application.NewRGB(27, 38, 54),
    Frameless:        false,
    MinWidth:         400,
    MinHeight:        300,
    Mac: application.MacWindow{
        InvisibleTitleBarHeight: 50,
        Backdrop:                application.MacBackdropTranslucent,
        TitleBar:                application.MacTitleBarHiddenInset,
    },
})
```

### Multi-Window Support

Wails v3 fully supports multiple windows. Each window can have different URLs and configurations:

```go
// Main window
app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title: "Main",
    URL:   "/",
})

// Settings window
app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title: "Settings",
    URL:   "/settings",
    Width: 600,
    Height: 400,
})
```

### Window Control Methods (Go)

```go
win.Show()
win.Hide()
win.Close()
win.Minimise()
win.Maximise()
win.Fullscreen()
win.SetTitle("New Title")
win.SetSize(800, 600)
win.SetPosition(100, 100)
win.Centre()
```

---

## Dialogs

### Message Dialogs

```go
// Information dialog
app.Dialog.Info().
    SetTitle("Info").
    SetMessage("Operation completed").
    Show()

// Warning dialog
app.Dialog.Warning().
    SetTitle("Warning").
    SetMessage("Are you sure?").
    Show()

// Error dialog
app.Dialog.Error().
    SetTitle("Error").
    SetMessage("Something went wrong").
    Show()

// Question dialog
app.Dialog.Question().
    SetTitle("Confirm").
    SetMessage("Delete this item?").
    Show()
```

### File Dialogs

```go
// Open file
path := app.Dialog.OpenFile()

// Save file
path := app.Dialog.SaveFile()
```

---

## Menus

```go
// Create an application menu
menu := app.Menu.New()

// Add items
fileMenu := menu.AddSubmenu("File")
fileMenu.Add("New").OnClick(func(ctx *application.Context) {
    // handle new
})
fileMenu.Add("Open").OnClick(func(ctx *application.Context) {
    // handle open
})
fileMenu.AddSeparator()
fileMenu.Add("Quit").OnClick(func(ctx *application.Context) {
    app.Quit()
})

// Set as application menu
app.Menu.SetApplicationMenu(menu)
```

---

## System Tray

```go
// Create system tray
tray := app.SystemTray.New()

// Set icon
tray.SetIcon(iconBytes)

// Attach a menu
trayMenu := app.Menu.New()
trayMenu.Add("Show").OnClick(func(ctx *application.Context) {
    win.Show()
})
trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
    app.Quit()
})
tray.SetMenu(trayMenu)

// Attach a window to the tray icon (toggle on click)
tray.AttachWindow(win)
```

---

## Drag & Drop and Frameless Windows

### File Drag & Drop

1. Enable in window options:
```go
app.Window.NewWithOptions(application.WebviewWindowOptions{
    EnableFileDrop: true,
})
```

2. Mark drop targets in HTML:
```html
<div data-file-drop-target>Drop files here</div>
```

3. Listen for events:
```go
win.OnWindowEvent(application.WindowFilesDropped, func(e *application.WindowEvent) {
    files := e.Data.([]string)
    // Handle dropped files
})
```

### Frameless Windows

```go
app.Window.NewWithOptions(application.WebviewWindowOptions{
    Frameless: true,
})
```

### Custom Draggable Title Bar (CSS)

When using frameless windows, define draggable regions with CSS:

```css
/* Make an element draggable (acts as title bar) */
.titlebar {
    --wails-draggable: drag;
}

/* Exclude interactive elements inside draggable regions */
.titlebar button {
    --wails-draggable: no-drag;
}
```

---

## Asset Embedding

### How Frontend Assets Are Embedded

```go
//go:embed all:frontend/dist
var assets embed.FS
```

The `all:` prefix includes dotfiles. The `frontend/dist` directory contains the built frontend.

### Configuring Asset Serving

```go
app := application.New(application.Options{
    Assets: application.AssetOptions{
        Handler: application.AssetFileServerFS(assets),
    },
})
```

### Development vs Production

- **`wails3 dev`**: Skips embedded assets, proxies to the Vite dev server (hot reload)
- **`wails3 build`**: Embeds `frontend/dist` into the Go binary (single executable)

### Custom Asset Handler

You can use any `http.Handler` for custom routing:

```go
Assets: application.AssetOptions{
    Handler: myCustomHandler, // any http.Handler
},
```

---

## CLI Commands

| Command | Description |
|---|---|
| `wails3 init -n myapp` | Create a new Wails project |
| `wails3 dev` | Run in dev mode (hot reload for both frontend and Go) |
| `wails3 build` | Build production binary |
| `wails3 generate bindings` | Generate JS/TS bindings from Go services |
| `wails3 task [taskname]` | Run a task defined in Taskfile.yml |
| `wails3 doctor` | Check system requirements |

### Install the CLI

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

---

## Build System (Taskfile)

Wails v3 uses [Task](https://taskfile.dev/) instead of Makefiles. The `Taskfile.yml` at the project root delegates to platform-specific taskfiles in `build/`.

### Common Tasks

```bash
wails3 task build              # Build for current platform
wails3 task dev                # Run in dev mode
wails3 task package            # Package for distribution
wails3 task build:server       # Build in server mode (no GUI)
```

### Passing Variables

```bash
wails3 task build PLATFORM=linux ARCH=amd64
```

---

## Platform-Specific Configuration

### macOS

```go
app := application.New(application.Options{
    Mac: application.MacOptions{
        ApplicationShouldTerminateAfterLastWindowClosed: true,
    },
})

// Window-level macOS options
app.Window.NewWithOptions(application.WebviewWindowOptions{
    Mac: application.MacWindow{
        InvisibleTitleBarHeight: 50,
        Backdrop:                application.MacBackdropTranslucent,
        TitleBar:                application.MacTitleBarHiddenInset,
    },
})
```

### Windows

```go
app := application.New(application.Options{
    Windows: application.WindowsOptions{
        // WebView2 settings, window class, etc.
        DisableQuitOnLastWindowClosed: false,
    },
})
```

### Window Button States (macOS & Windows only)

```go
app.Window.NewWithOptions(application.WebviewWindowOptions{
    MinimiseButtonState: application.ButtonEnabled,
    MaximiseButtonState: application.ButtonDisabled,
    CloseButtonState:    application.ButtonHidden,
})
```

---

## Common Patterns & Best Practices

### 1. Organize Services by Domain

```
cashier-v3/
├── main.go
├── services/
│   ├── product_service.go
│   ├── order_service.go
│   └── user_service.go
├── models/
│   ├── product.go
│   ├── order.go
│   └── user.go
```

### 2. Use Constructor Functions for Dependency Injection

```go
type OrderService struct {
    db     *sql.DB
    config *Config
}

func NewOrderService(db *sql.DB, config *Config) *OrderService {
    return &OrderService{db: db, config: config}
}
```

### 3. Return Errors Properly

Always return `error` as the last return value from service methods:

```go
func (s *OrderService) CreateOrder(items []OrderItem) (*Order, error) {
    if len(items) == 0 {
        return nil, fmt.Errorf("order must have at least one item")
    }
    // ...
}
```

### 4. Use Events for Real-Time Updates

For push-based updates from backend to frontend, use events instead of polling:

```go
// Backend pushes updates
go func() {
    for update := range updateChannel {
        app.Event.Emit("inventory-update", update)
    }
}()
```

```typescript
// Frontend listens
Events.On("inventory-update", (event) => {
    updateInventoryUI(event.data);
});
```

### 5. Use `data-wml-*` Attributes for Common Actions

Wails provides HTML attributes for common operations without writing JS:

```html
<!-- Open URL in system browser -->
<a data-wml-openURL="https://example.com">Visit Site</a>
```

### 6. Vite Config Must Include Wails Plugin

```typescript
// vite.config.ts
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
    plugins: [vue(), wails("./bindings")],
    server: {
        host: "127.0.0.1",
        port: Number(process.env.WAILS_VITE_PORT) || 9245,
        strictPort: true,
    },
});
```

---

## Critical Mistakes to Avoid

### ❌ DO NOT: Edit files in `frontend/bindings/`
These are auto-generated. They will be overwritten on next build or binding generation.

### ❌ DO NOT: Use Wails v2 API patterns
```go
// WRONG (v2 pattern)
wails.Run(&options.App{
    Bind: []interface{}{&MyStruct{}},
})

// CORRECT (v3 pattern)
app := application.New(application.Options{
    Services: []application.Service{
        application.NewService(&MyStruct{}),
    },
})
app.Run()
```

### ❌ DO NOT: Use `window.runtime` (v2 pattern)
```typescript
// WRONG (v2)
window.runtime.EventsEmit("myevent");

// CORRECT (v3)
import { Events } from '@wailsio/runtime';
Events.Emit("myevent");
```

### ❌ DO NOT: Store context in service structs (v2 pattern)
```go
// WRONG (v2 pattern)
type MyService struct {
    ctx context.Context // DON'T store context as a field
}

// CORRECT (v3 pattern) — use ServiceStartup to receive context
func (s *MyService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
    // Use ctx here for initialization
    return nil
}
```

### ❌ DO NOT: Forget to register services
If your Go methods aren't appearing in frontend bindings, ensure the service is registered in `application.Options.Services`.

### ❌ DO NOT: Use unexported (lowercase) methods for frontend access
Only exported (uppercase) Go methods are exposed to the frontend:
```go
func (s *MyService) GetData() string { return "visible" }   // ✅ Accessible
func (s *MyService) getData() string { return "invisible" }  // ❌ Not accessible
```

### ❌ DO NOT: Forget to run `wails3 generate bindings` after adding new service methods
In dev mode this is automatic, but if bindings seem stale, regenerate them.

---

## This Project's Specifics

| Property | Value |
|---|---|
| **Go Module** | `changeme` (in `go.mod`) |
| **Wails Version** | `v3.0.0-alpha.95` |
| **Frontend Framework** | Vue 3 + TypeScript |
| **Build Tool** | Vite 8 |
| **Vite Port** | `9245` (configurable via `WAILS_VITE_PORT`) |
| **Package Manager** | npm |
| **Frontend Entry** | `frontend/src/main.ts` |
| **Root Component** | `frontend/src/App.vue` |
| **Bindings Location** | `frontend/bindings/changeme/` |
| **Runtime Package** | `@wailsio/runtime` (installed in frontend) |

### Dev Workflow

```bash
# Start development (auto-reloads frontend + backend)
wails3 dev

# Or using Task
wails3 task dev

# Generate bindings manually (usually auto in dev mode)
wails3 generate bindings

# Production build
wails3 task build
```

### Adding a New Service

1. Create a Go file with your service struct and exported methods
2. Register it in `main.go` under `application.Options.Services`
3. Run `wails3 generate bindings` (or restart dev mode)
4. Import the generated binding in your Vue component
5. Call the method as an async function

### Frontend Binding Import Pattern

```typescript
// Import from the bindings directory matching your Go module name
import { MethodName } from '../bindings/changeme/ServiceName';

// Call as async
const result = await MethodName(arg1, arg2);
```
