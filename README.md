# Finance Dashboard API

A backend REST API for a multi-user finance dashboard system. Built with Go, Chi, and PostgreSQL — no ORM, raw SQL only. Supports role-based access control, JWT authentication, financial record management, and aggregated dashboard analytics.

---

## Table of Contents

- [Stack](#stack)
- [Project Structure](#project-structure)
- [Design Decisions](#design-decisions)
- [Database Schema](#database-schema)
- [Setup & Running](#setup--running)
- [API Reference](#api-reference)
- [Role Permissions](#role-permissions)
- [Assumptions](#assumptions)
- [Seeded Test Accounts](#seeded-test-accounts)
- [Performance](#performance)

---

## Stack

| Layer | Choice | Why |
|---|---|---|
| Language | Go | Systems-level control, strong concurrency, explicit error handling |
| Router | Chi | Lightweight, idiomatic, uses standard `net/http` — no magic |
| Database | PostgreSQL | Aggregation queries (`DATE_TRUNC`, `GROUP BY`) are native and fast |
| DB Driver | pgx v5 | Raw SQL, no ORM — full control over every query |
| Auth | JWT (HS256) | Stateless, role encoded in token — no DB hit on every request |
| Password | bcrypt | Adaptive hashing, resistant to brute force |
| Migrations | golang-migrate | Plain `.sql` files, easy to reason about and version control |

---

## Project Structure

```
finance-dashboard/
├── cmd/
│   ├── server/
│   │   └── main.go                 # entry point — wires everything, starts server
│   └── seed/
│       └── main.go                 # seeds 1000 transactions + 3 test users
│
├── internal/
│   ├── config/
│   │   └── config.go               # loads env vars, crashes fast if required vars missing
│   ├── db/
│   │   └── db.go                   # pgxpool setup with ping check on startup
│   ├── dto/
│   │   └── dto.go                  # input/filter structs shared across layers
│   ├── middleware/
│   │   ├── auth.go                 # JWT validation, injects user_id + role into context
│   │   ├── role.go                 # RequireRole(...) — reads role from context, returns 403
│   │   └── logger.go               # logs method, path, status, latency on every request
│   ├── model/
│   │   ├── user.go                 # User struct, Role type + constants
│   │   ├── transactions.go         # Transaction struct, TxnType constants
│   │   └── share.go                # SharedAccess struct, ShareRequest
│   ├── repository/
│   │   ├── user_repo.go            # all user SQL queries
│   │   ├── transaction_repo.go     # all transaction SQL queries + aggregations
│   │   └── share_repo.go           # shared access grant queries
│   ├── service/
│   │   ├── auth_service.go         # register, login, bcrypt, JWT generation
│   │   ├── user_service.go         # list users, update role, toggle status
│   │   └── transaction_service.go  # business logic wrapping transaction repo
│   ├── handler/
│   │   ├── auth_handler.go         # POST /auth/register, POST /auth/login
│   │   ├── user_handler.go         # user management (admin only)
│   │   ├── transaction_handler.go  # transaction CRUD + dashboard endpoints
│   │   ├── share_handler.go        # POST /share
│   │   └── helper.go               # writeJSON, writeError, parseIntQuery, handleRepoError
│   ├── router/
│   │   └── router.go               # all routes with middleware attached
│   └── validator/
│       └── validator.go            # go-playground/validator wrapper with human-readable errors
│
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000002_create_transactions.up.sql
│   └── 000003_create_shared_access.up.sql
│
├── .env.example
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## Design Decisions

### No ORM — raw SQL with pgx

GORM and Ent abstract away SQL in ways that make aggregation queries awkward. The dashboard endpoints use `DATE_TRUNC`, `COALESCE`, `GROUP BY`, and conditional `SUM` — these are cleaner and more explicit as raw SQL. pgx v5 maps rows to structs directly with `.Scan()`, which is all that's needed here.

### `dto` package to break import cycles

Services need input structs, repositories also need those same structs to build queries. If both `service` and `repository` import each other, Go refuses to compile. Moving `CreateTransactionInput`, `UpdateTransactionInput`, and `TransactionFilter` into a standalone `internal/dto` package breaks the cycle cleanly:

```
handler → service → repository
    ↘        ↘         ↘
            dto        dto
```

### Role encoded in JWT

On login, the user's role is embedded in the JWT claims. The `RequireRole` middleware reads it directly from context — no database query needed per request. This keeps protected endpoints fast.

### `contextKey` typed string for context values

Instead of using raw strings as context keys (`"user_id"`, `"role"`), a typed `contextKey` type is used. This prevents accidental collisions with other packages that might use the same string keys, which is a real issue in middleware-heavy Go applications.

### Soft delete on transactions

Transactions are never hard deleted. Setting `deleted_at` to the current timestamp marks them as deleted, and all queries filter `WHERE deleted_at IS NULL`. This preserves history, makes auditing possible, and avoids accidental data loss.

### `NUMERIC(12,2)` not `FLOAT` for money

Floating point types (`FLOAT`, `DOUBLE`) cannot represent decimal fractions exactly. `0.1 + 0.2` in floating point is `0.30000000000000004`. For financial data, this is unacceptable. `NUMERIC(12,2)` stores exact decimal values in Postgres.

### Admin sees global data, others see own data

On all dashboard and listing endpoints, admins see data across all users. Analysts and viewers are scoped to their own transactions. This is enforced at the repository layer — the `userID` filter is applied or omitted based on role, not in the handler.

### Shared access feature

Users can share their transaction data with other users at a `viewer` or `analyst` permission level via `POST /share`. The `shared_access` table tracks `owner_id → shared_with_id` pairs. Queries for non-admin users include:

```sql
WHERE user_id = $1
   OR user_id IN (SELECT owner_id FROM shared_access WHERE shared_with_id = $1)
```

This lets a viewer or analyst see both their own data and data shared with them.

### Crash fast on missing config

`config.go` calls `log.Fatalf` if `DATABASE_URL` or `JWT_SECRET` are not set. The server refuses to start with a clear error rather than starting and failing with a cryptic panic deep in a handler.

### Indexes on transactions

Four indexes are created on the `transactions` table:

```sql
CREATE INDEX idx_transactions_user_id  ON transactions(user_id);
CREATE INDEX idx_transactions_date     ON transactions(date);
CREATE INDEX idx_transactions_type     ON transactions(type);
CREATE INDEX idx_transactions_category ON transactions(category);
```

The dashboard aggregation queries filter and group on `date`, `type`, and `category`. Without indexes, those queries do full table scans. With 1000+ rows, the difference is measurable.

### `main.go` is only wiring

`main.go` does exactly four things: load config, connect DB, wire dependencies (repo → service → handler), start server. No business logic lives there. If the structure needs to change, the change location is obvious.

---

## Database Schema

### users

```sql
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    password   TEXT NOT NULL,           -- bcrypt hash
    name       TEXT NOT NULL,
    role       user_role NOT NULL DEFAULT 'viewer',
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### transactions

```sql
CREATE TABLE transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    type        txn_type NOT NULL,       -- 'income' | 'expense'
    category    TEXT NOT NULL,
    description TEXT,
    date        DATE NOT NULL,
    deleted_at  TIMESTAMPTZ,             -- NULL = active, set = soft deleted
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### shared_access

```sql
CREATE TABLE shared_access (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission     share_permission NOT NULL,  -- 'viewer' | 'analyst'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(owner_id, shared_with_id)
);
```

---

## Setup & Running

### Prerequisites

- Go 1.21+
- Docker
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Steps

```bash
# 1. Clone and install dependencies
git clone https://github.com/PratikkJadhav/Finance-API
cd Finance-API
go mod tidy

# 2. Set up environment
cp .env.example .env
# Edit .env if needed

# 3. Start Postgres
make docker-up

# 4. Run migrations
make migrate-up

# 5. Seed test data (creates 3 users + 1000 transactions)
make seed

# 6. Start server
make run
```

### Makefile Commands

| Command | Description |
|---|---|
| `make run` | Start the server |
| `make build` | Build binary to `bin/server` |
| `make docker-up` | Start Postgres container |
| `make docker-down` | Stop Postgres container |
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Roll back last migration |
| `make seed` | Seed 3 users + 1000 transactions |

---

## API Reference

All protected endpoints require `Authorization: Bearer <token>` header.

### Auth

#### `POST /auth/register`

```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "role": "analyst"
}
```

Role defaults to `viewer` if omitted. Valid roles: `viewer`, `analyst`, `admin`.

**Response `201`:**
```json
{
  "message": "user registered successfully",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "role": "analyst",
    "is_active": true,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

---

#### `POST /auth/login`

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response `200`:**
```json
{
  "token": "eyJhbGci...",
  "user": { ... }
}
```

---

### Transactions

#### `GET /transactions`

Query params (all optional):

| Param | Example | Description |
|---|---|---|
| `type` | `income` | Filter by type |
| `category` | `salary` | Filter by category |
| `from` | `2024-01-01` | Date range start |
| `to` | `2024-12-31` | Date range end |
| `page` | `1` | Page number (default: 1) |
| `limit` | `20` | Items per page (default: 20) |

**Response `200`:**
```json
{
  "transactions": [...],
  "total": 142,
  "page": 1,
  "limit": 20
}
```

---

#### `GET /transactions/{id}`

**Response `200`:** Single transaction object. Returns `404` if not found or soft deleted.

---

#### `POST /transactions` _(analyst, admin)_

```json
{
  "amount": 1500.00,
  "type": "income",
  "category": "salary",
  "description": "Monthly salary",
  "date": "2024-01-15"
}
```

**Response `201`:** Created transaction object.

---

#### `PUT /transactions/{id}` _(analyst, admin)_

Same body as POST. Replaces all fields.

**Response `200`:** Updated transaction object.

---

#### `DELETE /transactions/{id}` _(admin only)_

Soft delete — sets `deleted_at`, data is not removed.

**Response `200`:**
```json
{ "message": "transaction deleted successfully" }
```

---

### Dashboard

#### `GET /dashboard/summary`

Returns totals across all non-deleted transactions. Admins see global totals, others see their own (including shared).

**Response `200`:**
```json
{
  "total_income": 45230.50,
  "total_expense": 18940.75,
  "net_balance": 26289.75
}
```

---

#### `GET /dashboard/recent`

Returns last 10 transactions ordered by `created_at DESC`.

**Response `200`:**
```json
{
  "recent": [...],
  "count": 10
}
```

---

#### `GET /dashboard/trends` _(analyst, admin)_

Monthly income and expense totals for the last 12 months.

**Response `200`:**
```json
{
  "trends": [
    { "month": "2024-12", "income": 5200.00, "expense": 2100.50 },
    { "month": "2024-11", "income": 4800.00, "expense": 1980.00 }
  ]
}
```

---

#### `GET /dashboard/categories` _(analyst, admin)_

Totals grouped by category and type, ordered by total descending.

**Response `200`:**
```json
{
  "categories": [
    { "category": "salary",  "type": "income",  "total": 18000.00 },
    { "category": "rent",    "type": "expense", "total": 6000.00 },
    { "category": "food",    "type": "expense", "total": 2400.00 }
  ]
}
```

---

### Users _(admin only)_

#### `GET /users`

**Response `200`:**
```json
{
  "users": [...],
  "count": 5
}
```

---

#### `PATCH /users/{id}/role`

```json
{ "role": "analyst" }
```

Valid roles: `viewer`, `analyst`, `admin`.

**Response `200`:**
```json
{ "message": "role updated successfully" }
```

---

#### `PATCH /users/{id}/status`

```json
{ "is_active": false }
```

Deactivated users cannot log in.

**Response `200`:**
```json
{ "message": "status updated successfully" }
```

---

### Sharing

#### `POST /share` _(any authenticated user)_

Share your transaction data with another user by their email.

```json
{
  "shared_with_email": "analyst@finance.com",
  "permission": "viewer"
}
```

Valid permissions: `viewer`, `analyst`. The target user will see the sharer's transactions in their own listing and dashboard responses.

**Response `200`:**
```json
{ "message": "Access granted successfully" }
```

---

## Role Permissions

| Action | Viewer | Analyst | Admin |
|---|:---:|:---:|:---:|
| Register / Login | ✓ | ✓ | ✓ |
| View own transactions | ✓ | ✓ | ✓ |
| View all transactions | ✗ | ✗ | ✓ |
| Create transactions | ✗ | ✓ | ✓ |
| Update transactions | ✗ | ✓ | ✓ |
| Delete transactions | ✗ | ✗ | ✓ |
| Dashboard summary | ✓ | ✓ | ✓ |
| Dashboard recent | ✓ | ✓ | ✓ |
| Dashboard trends | ✗ | ✓ | ✓ |
| Dashboard categories | ✗ | ✓ | ✓ |
| List all users | ✗ | ✗ | ✓ |
| Update user role | ✗ | ✗ | ✓ |
| Update user status | ✗ | ✗ | ✓ |
| Share data with others | ✓ | ✓ | ✓ |

---

## Assumptions

**Role assignment at registration** — a user's role is set when they register and can only be changed later by an admin via `PATCH /users/{id}/role`.

**Soft delete is permanent from the user's perspective** — once deleted, a transaction disappears from all listing and dashboard queries. It is not exposed via any endpoint. The data remains in the database for potential audit purposes.

**Viewers and analysts are data-scoped** — they only see transactions they own or that have been explicitly shared with them. Admins always see global data.

**JWT expiry is 24 hours** — after expiry the user must log in again. There is no refresh token mechanism.

**Amounts are always positive** — the `type` field (`income` or `expense`) determines the direction. A `CHECK (amount > 0)` constraint is enforced at the database level.

**Sharing is one-directional** — user A sharing with user B does not give user B the ability to create or modify user A's transactions. The `permission` field on `shared_access` controls whether B can only view or also create.

**Date format** — transaction dates should be sent as `YYYY-MM-DD` strings in request bodies.

---

## Seeded Test Accounts

After running `make seed`:

| Email | Password | Role |
|---|---|---|
| admin@finance.com | password123 | admin |
| analyst@finance.com | password123 | analyst |
| viewer@finance.com | password123 | viewer |

The seed script also inserts 1000 randomly generated transactions spread across the three users, covering the last 12 months with random amounts, types, and categories.

---

## Performance

Dashboard summary query aggregating 1000 transactions across all users:

```
GET /dashboard/summary  →  ~3ms  (local Postgres, seeded dataset)
GET /dashboard/trends   →  ~5ms  (12-month GROUP BY with DATE_TRUNC)
GET /dashboard/categories → ~4ms (GROUP BY category + type)
```

Indexes on `user_id`, `date`, `type`, and `category` keep aggregation queries fast as the dataset grows. The `deleted_at IS NULL` filter benefits from the same indexes since most rows will have `deleted_at = NULL`.
