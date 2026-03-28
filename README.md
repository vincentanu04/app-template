# app-template

Full-stack app template: **Go** backend (Chi router, go-jet, Goose, oapi-codegen) + **React/TypeScript** frontend (Vite, Tailwind v4, RTK Query, shadcn/ui).

## Tech stack

| Layer | Tool |
|---|---|
| Router | [Chi v5](https://github.com/go-chi/chi) |
| DB queries | [go-jet v2](https://github.com/go-jet/jet) |
| Migrations | [Goose v3](https://github.com/pressly/goose) |
| API codegen | [oapi-codegen v2](https://github.com/oapi-codegen/oapi-codegen) |
| Frontend state | [RTK Query](https://redux-toolkit.js.org/rtk-query/overview) |
| UI components | [shadcn/ui](https://ui.shadcn.com) + Tailwind v4 |

---

## Getting started

### 1. Clone and configure

```bash
cp .env.example .env.local
# Edit .env.local with your values
```

### 2. Start the database

```bash
docker compose up -d
```

### 3. Run migrations and generate DB models

```bash
cd backend
make migration-up
```

This runs Goose migrations then regenerates go-jet type-safe models into `backend/generated/db/`.

### 4. Generate the Go server and TypeScript client from OpenAPI

```bash
make openapi-codegen
```

This generates:
- `backend/generated/server/server.gen.go` — Chi server + strict interfaces
- `frontend/src/api/client.ts` — RTK Query hooks

### 5. Run the backend

```bash
make run
# or with hot reload (requires air):
# air
```

### 6. Run the frontend

```bash
cd ../frontend
npm install
npm run dev
```

---

## Project structure

```
.
├── docker-compose.yml          # Local Postgres
├── Dockerfile                  # Multi-stage production build
├── .env.example
├── backend/
│   ├── Makefile                 # Dev workflow commands
│   ├── main.go                 # Entry point
│   ├── go.mod
│   ├── api/
│   │   ├── openapi.yml         # OpenAPI spec (source of truth)
│   │   └── oapi-server.yml     # oapi-codegen config
│   ├── db/
│   │   └── migrations/         # Goose SQL migrations
│   ├── generated/
│   │   ├── db/                 # go-jet generated models (auto)
│   │   └── server/             # oapi-codegen generated server (auto)
│   └── internal/
│       ├── db/postgres.go      # DB connection
│       ├── deps/deps.go        # Dependency container
│       ├── server/             # HTTP handlers + middleware
│       └── app/
│           ├── repository/     # go-jet SELECT queries
│           └── mutation/       # go-jet INSERT/UPDATE/DELETE queries
└── frontend/
    ├── rtk-query.config.cjs    # RTK Query codegen config
    ├── src/
    │   ├── store/api.ts        # RTK Query base API
    │   ├── store/index.ts      # Redux store
    │   └── api/client.ts       # Generated RTK hooks (auto)
    └── ...
```

---

## Development workflow

### Adding a new API endpoint

1. Add the endpoint to `backend/api/openapi.yml`
2. Run `make openapi-codegen` to regenerate server + TS client
3. Implement the handler method in `backend/internal/server/handlers.go`
4. Add repository/mutation functions in `backend/internal/app/`

### Adding a database table

1. Create a migration: `make migration-create name=add_my_table`
2. Edit the generated SQL file in `backend/db/migrations/`
3. Run `make migration-up` to apply + regenerate go-jet models
4. Use the generated types from `myapp/generated/db/myapp_db/public/model` and `table`

### go-jet query patterns

```go
// SELECT
stmt := pg.SELECT(tbl.AllColumns).FROM(tbl).WHERE(tbl.ID.EQ(pg.UUID(id)))
var rows []model.ItemTbl
stmt.QueryContext(ctx, db, &rows)

// INSERT with RETURNING
stmt := tbl.INSERT(tbl.MutableColumns).MODEL(record).RETURNING(tbl.AllColumns)
var dest []model.ItemTbl
stmt.QueryContext(ctx, db, &dest)

// UPDATE
stmt := tbl.UPDATE(tbl.Name, tbl.UpdatedAt).SET(name, pg.TimestampzT(now)).
    WHERE(tbl.ID.EQ(pg.UUID(id))).RETURNING(tbl.AllColumns)

// DELETE
stmt := tbl.DELETE().WHERE(tbl.ID.EQ(pg.UUID(id)))
stmt.ExecContext(ctx, db)
```

---

## Renaming the module

To use a custom module name instead of `myapp`:

1. Update `backend/go.mod`: change `module myapp` to `module github.com/yourname/yourapp`
2. Update all internal import paths (find/replace `myapp/` → `github.com/yourname/yourapp/`)
3. Update the generated DB import path in `repository/` and `mutation/` files to match your database name (replace `myapp_db`)
4. Run `make tidy`