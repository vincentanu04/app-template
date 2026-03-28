# App Template — AI Coding Agent Instructions

## Project Overview

Full-stack web application template with a **Go backend** (Chi router, PostgreSQL, go-jet) and **React/TypeScript frontend** (Vite, RTK Query, shadcn/ui, Tailwind v4).

**Architecture**: Monorepo with `backend/` and `frontend/` directories. Backend uses OpenAPI-first design with code generation. Frontend consumes auto-generated RTK Query hooks.

**Local Development**:
- Run via Docker Compose (all 3 services: postgres, backend with air hot-reload, frontend with Vite dev server)
- Or run individually: `make run` in `backend/`, `npm run dev` in `frontend/`
- Environment variables in `.env.local` at repo root (loaded by Makefile via `-include ../.env.local`)

**Go Module Name**: `myapp` — update this in `go.mod`, `main.go`, and all import paths when forking the template.

---

## Critical Development Workflows

### Backend: OpenAPI-First Development
**All API changes must follow this exact sequence:**

1. **Edit OpenAPI spec** (`backend/api/openapi.yml`)
   - Define new paths, request/response schemas
   - Mark public routes with `security: []`

2. **Generate code**: `cd backend && make openapi-codegen`
   - Generates `generated/server/generated.go` — strict server interface + Chi router wiring
   - Generates `frontend/src/api/client.ts` — RTK Query hooks (do NOT edit manually)

3. **Implement the handler** in `internal/handler/`
   - One file per endpoint, named after the operation (e.g. `get_auth_me.go`)
   - Handler calls into `internal/app_service/` for business logic

4. **Routes are auto-wired** by `oapi.HandlerFromMux(h, apiRouter)` in `main.go`

**Example handler** (auth-protected endpoint):
```go
func (h *Handler) GetAuthMe(ctx context.Context, _ oapi.GetAuthMeRequestObject) (oapi.GetAuthMeResponseObject, error) {
    userID := middleware.UserIDFromContext(ctx)

    user, err := repository.GetUserByID(ctx, h.deps.DB, userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return oapi.GetAuthMe401JSONResponse{Message: "unauthorized"}, nil
    }

    return oapi.GetAuthMe200JSONResponse{Id: user.ID, Email: user.Email}, nil
}
```

**Example handler** (custom response that sets a cookie — not managed by oapi responses):
```go
type loginCookieResponse struct {
    user  oapi.AuthUser
    token string
}

func (r loginCookieResponse) VisitPostAuthLoginResponse(w http.ResponseWriter) error {
    http.SetCookie(w, &http.Cookie{Name: middleware.CookieName, Value: r.token, ...})
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    return json.NewEncoder(w).Encode(r.user)
}

func (h *Handler) PostAuthLogin(ctx context.Context, request oapi.PostAuthLoginRequestObject) (oapi.PostAuthLoginResponseObject, error) {
    user, token, err := auth.Login(ctx, h.deps, string(request.Body.Email), request.Body.Password)
    // ...
    return loginCookieResponse{user: ..., token: token}, nil
}
```

### Backend: Database Migrations
**Pattern**: Sequential Goose migrations + auto-generated type-safe models (go-jet)

1. **Create migration**: `make migration-create name=<migration_name>`
   - Creates a sequential `.sql` file in `db/migrations/`
   - Edit both `-- +goose Up` and `-- +goose Down` blocks

2. **Run migrations**: `make migration-up`
   - Migrates local DB using `DATABASE_URL` from `.env.local`
   - Auto-regenerates go-jet models in `generated/db/`

3. **Rollback**: `make migration-down`

4. **Status**: `make migration-status`

**Migration file format**:
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE example_tbl (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS example_tbl;
-- +goose StatementEnd
```

**Never write raw SQL in business logic** — use go-jet models from `generated/db/myapp_db/public/model` and table structs from `generated/db/myapp_db/public/table`.

### Backend: Repository + Mutation Pattern

**Repository** = read-only queries (`internal/repository/`)
**Mutation** = write operations (`internal/mutation/`)

Both accept `qrm.DB` (accepts `*sql.DB` or `*sql.Tx`).

**Repository example**:
```go
package repository

import (
    pg "github.com/go-jet/jet/v2/postgres"
    "github.com/go-jet/jet/v2/qrm"
    "myapp/generated/db/myapp_db/public/model"
    "myapp/generated/db/myapp_db/public/table"
)

func GetUserByEmail(ctx context.Context, db qrm.DB, email string) (*model.UserTbl, error) {
    tbl := table.UserTbl

    stmt := pg.SELECT(tbl.AllColumns).
        FROM(tbl).
        WHERE(tbl.Email.EQ(pg.String(email)))

    var rows []model.UserTbl
    if err := stmt.QueryContext(ctx, db, &rows); err != nil {
        return nil, err
    }
    if len(rows) == 0 {
        return nil, nil
    }
    return &rows[0], nil
}
```

**Mutation example**:
```go
func InsertUser(ctx context.Context, db qrm.DB, email, passwordHash string) (*model.UserTbl, error) {
    tbl := table.UserTbl

    record := model.UserTbl{
        ID:           uuid.New(),
        Email:        email,
        PasswordHash: passwordHash,
        CreatedAt:    time.Now().UTC(),
    }

    stmt := tbl.INSERT(tbl.MutableColumns).
        MODEL(record).
        RETURNING(tbl.AllColumns)

    var dest []model.UserTbl
    if err := stmt.QueryContext(ctx, db, &dest); err != nil {
        return nil, err
    }
    return &dest[0], nil
}
```

**go-jet gotchas**:
- `tbl.UPDATE(columnList)` — pass `pg.ColumnList` directly, never spread (`...`)
- Dynamic updates: build a `pg.ColumnList` and corresponding model fields conditionally
- `tbl.MutableColumns` excludes the primary key; use for INSERT
- Always `RETURNING(tbl.AllColumns)` on INSERT/UPDATE to get the DB-assigned values back

### Backend: App Service Layer

`internal/app_service/<domain>/` — one file per action (e.g. `login.go`, `register.go`)

- Orchestrates repository reads, mutation writes, and side-effects (JWT generation, email, etc.)
- Receives `deps.Deps` (not raw `*sql.DB`) so it can access any future dependency
- Returns domain errors (`errors.New("...")`) for expected failures; return raw `error` for unexpected ones
- Handlers switch on error type/value to map to HTTP response codes

**App service example**:
```go
package auth

var ErrInvalidCredentials = errors.New("invalid credentials")

func Login(ctx context.Context, d deps.Deps, email, password string) (*model.UserTbl, string, error) {
    user, err := repository.GetUserByEmail(ctx, d.DB, email)
    if err != nil {
        return nil, "", err
    }
    if user == nil {
        return nil, "", ErrInvalidCredentials
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, "", ErrInvalidCredentials
    }
    token, err := GenerateToken(user.ID)
    if err != nil {
        return nil, "", err
    }
    return user, token, nil
}
```

### Backend: Authentication

- JWT stored in an `HttpOnly` cookie named `access_token` (see `middleware.CookieName`)
- Cookie is set on login/register, cleared on logout
- Middleware (`internal/middleware/auth.go`) validates the token on every request
- Public routes are bypassed in the switch statement at the top of `Auth()`:
  ```go
  switch r.URL.Path {
  case "/api/auth/register", "/api/auth/login", "/api/auth/logout", "/api/health":
      next.ServeHTTP(w, r)
      return
  }
  ```
- To add a new public route, add its path to that switch case
- User ID is available in handlers via `middleware.UserIDFromContext(ctx)`

### Backend: Adding Dependencies

- Add application dependencies (DB, Redis, etc.) to `internal/deps/deps.go`
- Pass `deps.Deps` through handler → app_service (never pass `*sql.DB` directly to handlers)
- Go tool dependencies (codegen, migration runner): add to `go.mod` `tool` block, then `go mod tidy`

### Backend: Running Locally

```bash
# With hot reload (recommended)
cd backend && make run       # uses air — auto-reloads on .go file changes

# Without hot reload
cd backend && make run-once  # go run main.go

# Full dev environment (DB + backend + frontend)
docker compose up -d         # from repo root
```

**Ports**: backend on `:8080`, frontend dev server on `:5173`

### Backend: Key Files Reference

| File | Purpose |
|------|---------|
| `main.go` | Chi router setup, middleware, SPA fallback, server start |
| `api/openapi.yml` | Source of truth for all API endpoints and schemas |
| `api/oapi-server.yml` | oapi-codegen config → generates `generated/server/generated.go` |
| `Makefile` | All dev commands (codegen, migrations, run, fmt) |
| `air.toml` | Hot reload config — watches `.go` and `.env.local` |
| `internal/handler/handler.go` | `Handler` struct + `NewHandler` constructor |
| `internal/deps/deps.go` | Shared dependency container |
| `internal/middleware/auth.go` | JWT cookie auth + public route bypass |
| `internal/middleware/rate_limit.go` | Per-IP rate limiter (100 req/min) |
| `db/migrations/` | Goose SQL migration files |
| `generated/server/generated.go` | Auto-generated — do NOT edit |
| `generated/db/` | Auto-generated go-jet models — do NOT edit |

---

## Frontend: Architecture

**Tech stack**: React 19, TypeScript, Vite 6, Tailwind v4, RTK Query, shadcn/ui, react-router-dom v7

**Key principle**: All API hooks are auto-generated — only write page/component code that calls those hooks.

### Frontend: Type-Safe API Calls

**Generated hooks** live in `src/api/client.ts` — do NOT edit this file manually. Regenerate via:
```bash
cd backend && make openapi-codegen
# or from frontend/:
npm run codegen
```

**Base API** (`src/store/api.ts`):
- `baseUrl: '/api'` — relative, handled by Vite proxy in dev
- `credentials: 'include'` — sends cookies with every request (required for JWT auth)
- Add new `tagTypes` here when adding resource-specific cache invalidation

**RTK Query config** (`frontend/rtk-query.config.cjs`):
- `schemaFile` points to `../backend/api/openapi.yml`
- `outputFile` → `src/api/client.ts`

**Using hooks**:
```tsx
import { useGetAuthMeQuery, usePostAuthLogoutMutation } from '@/api/client'

const MyComponent = () => {
  const { data: me, isLoading } = useGetAuthMeQuery()
  const [logout] = usePostAuthLogoutMutation()
  // ...
}
```

### Frontend: Adding a New Page

1. Add the route to `backend/api/openapi.yml` and run `make openapi-codegen` (backend)
2. Create `frontend/src/pages/MyPage.tsx`
3. Add the route to `src/App.tsx` — wrap in `<ProtectedRoute>` if auth is required

**Page component pattern** (always use arrow functions):
```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSomeMutation } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const MyPage = () => {
  const navigate = useNavigate()
  const [doThing, { isLoading }] = useSomeMutation()

  const handleClick = async () => {
    await doThing({ ... })
    navigate('/somewhere')
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Page Title</CardTitle>
        </CardHeader>
        <CardContent>
          <Button onClick={handleClick} disabled={isLoading}>Go</Button>
        </CardContent>
      </Card>
    </div>
  )
}

export default MyPage
```

### Frontend: Protected Routes

`src/components/ProtectedRoute.tsx` — uses `useGetAuthMeQuery()` to gate access:
- While loading: shows a spinner
- On error (401) or no data: redirects to `/login`
- Wrap any private route in `<ProtectedRoute>` in `App.tsx`

```tsx
<Route
  path="/dashboard"
  element={
    <ProtectedRoute>
      <DashboardPage />
    </ProtectedRoute>
  }
/>
```

### Frontend: UI Components (shadcn/ui)

Components live in `src/components/ui/`. Install new ones with:
```bash
cd frontend && npx --yes shadcn@latest add <component-name>
```

**Always use shadcn components over raw HTML**:
- `Button` — with `variant` (`default`, `outline`, `destructive`, `ghost`) and `size` (`sm`, `default`, `lg`)
- `Input` — controlled inputs
- `Label` — form labels (pair with `htmlFor`)
- `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent` — layout containers
- `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableHead`, `TableCell` — data tables

**`cn()` utility** (`src/lib/utils.ts`): merges Tailwind classes safely — always use it when combining class strings:
```tsx
import { cn } from '@/lib/utils'
<div className={cn('base-class', isActive && 'active-class', className)} />
```

### Frontend: Styling Conventions

- Use Tailwind v4 utility classes — no inline styles
- Use CSS variables for theming: `bg-muted`, `text-muted-foreground`, `text-destructive`, `bg-background`, etc.
- Spacing: use Tailwind scale (`p-4`, `gap-2`, `mt-6`) rather than custom pixel values
- Always arrow functions for components and handlers — no `function` declarations

### Frontend: State Management

- **API state**: RTK Query (auto-generated hooks) — do not duplicate with useState
- **Redux slices** (`src/store/`): only for global UI state not covered by RTK Query
- **Local state** (`useState`): form fields, UI toggles within a single component
- Redux store is set up in `src/store/index.ts` — add slice reducers there

### Frontend: Key Files Reference

| File | Purpose |
|------|---------|
| `src/App.tsx` | Route definitions — add all new routes here |
| `src/main.tsx` | App entry point — Redux `Provider` wraps the tree |
| `src/api/client.ts` | Auto-generated RTK Query hooks — do NOT edit |
| `src/store/api.ts` | Base RTK Query API (baseUrl, credentials, tagTypes) |
| `src/store/index.ts` | Redux store + middleware config |
| `src/components/ProtectedRoute.tsx` | Auth gate component |
| `src/components/ui/` | shadcn/ui components |
| `src/lib/utils.ts` | `cn()` Tailwind merge helper |
| `src/pages/` | One file per page/route |
| `vite.config.ts` | Vite config — includes `/api` proxy to backend |
| `rtk-query.config.cjs` | RTK Query codegen config |
| `components.json` | shadcn/ui config (style, aliases, icon library) |

---

## Docker Compose (Dev Environment)

```yaml
services:
  myapp-db:       # Postgres 17, port 5432, health-checked
  myapp-backend:  # golang:1.25 image, mounts ./backend, runs air
  myapp-frontend: # node:22 image, mounts ./frontend, runs Vite
```

**Backend startup sequence** (in compose command):
1. `goose up` — run pending migrations
2. `jet codegen` — regenerate go-jet models
3. `air` — start with hot reload

**Vite proxy**: All `/api` requests from the browser are proxied to `myapp-backend:8080` via Vite dev server. In production, the Go server itself serves the built frontend and handles `/api` routes.

---

## Common Pitfalls

1. **Never edit generated files**: `generated/server/generated.go`, `frontend/src/api/client.ts`, `generated/db/`
2. **Always run codegen after OpenAPI changes**: both backend and frontend types regenerate together — `make openapi-codegen`
3. **`types.Email` from oapi-codegen**: cast to `string` before passing to functions — `string(request.Body.Email)`
4. **go-jet `UPDATE` takes `ColumnList`, not a spread**: use `tbl.UPDATE(cols)` not `tbl.UPDATE(cols...)`
5. **JWT auth cookie requires `credentials: 'include'`** on the frontend base query — already set in `store/api.ts`
6. **New public routes must be added to the bypass list** in `internal/middleware/auth.go`
7. **`lib/utils.ts` must exist** for shadcn components to compile — it exports `cn()` using `clsx` + `tailwind-merge`
8. **Migration numbers are sequential** — never skip or reuse a number; name format is `00001_description.sql`
9. **After `make migration-up`, go-jet models regenerate automatically** — you must commit both the migration and the generated `generated/db/` changes

---

## Adding a New Feature (End-to-End Checklist)

1. [ ] Add migration if new table needed: `make migration-create name=<name>` → edit SQL → `make migration-up`
2. [ ] Add repository function in `internal/repository/`
3. [ ] Add mutation function in `internal/mutation/` (if writes needed)
4. [ ] Add app service in `internal/app_service/<domain>/<action>.go`
5. [ ] Add OpenAPI path + schemas in `backend/api/openapi.yml`
6. [ ] Run `make openapi-codegen` — generates handler stub + frontend hook
7. [ ] Implement handler in `internal/handler/<operation>.go`
8. [ ] If public route: add path to bypass list in `internal/middleware/auth.go`
9. [ ] Create frontend page in `frontend/src/pages/`
10. [ ] Add route in `frontend/src/App.tsx` (wrap in `<ProtectedRoute>` if private)
11. [ ] Run `go build ./...` to verify backend compiles
