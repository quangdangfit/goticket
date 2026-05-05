# GoTicket — High-Load Event Ticketing Platform

A Go service for selling tickets to events / concerts / shows, designed for
high concurrency (flash-sale scale: 10k+ RPS during on-sale moments) and
correctness under contention (no oversells, no double-charges).

Architecture: **Domain-Driven Design**. Each bounded context lives in its
own folder under `internal/<domain>/`, with a ports-and-adapters layout
inspired by [goshop](https://github.com/quangdangfit/goshop).

> All code must follow the rules in `.claude/rules/`:
> `coding-style.md`, `dependency-injection.md`, `pre-commit-checks.md`,
> `security.md`, `testing.md`. Use `/implement-phase <N>` to execute a
> phase, `/endpoint <desc>` to add a new HTTP endpoint.

---

## 1. Tech stack

| Concern              | Choice                                          |
|----------------------|-------------------------------------------------|
| Language             | Go 1.25+                                        |
| HTTP framework       | Gin                                             |
| ORM                  | GORM v2 (MySQL driver)                          |
| Primary DB           | MySQL 8 (InnoDB, READ-COMMITTED)                |
| Cache / locks / hold | Redis 7 (single + cluster ready)                |
| Async pipeline       | Kafka (segmentio/kafka-go)                      |
| Migrations           | golang-migrate                                  |
| Auth                 | JWT (access + refresh)                          |
| Validation           | go-playground/validator                         |
| Config               | viper (`config/config.yaml` + env override)     |
| Logger               | `log/slog` JSON handler                         |
| Mocks                | go.uber.org/mock (`mockgen`)                    |
| Integration tests    | testcontainers-go                               |
| Observability        | OpenTelemetry traces + Prometheus metrics       |
| Container            | Docker + docker-compose for local dev           |

---

## 2. Repository layout

```
goticket/
├── cmd/
│   └── server/main.go            # Single composition root (DI wiring)
├── config/
│   ├── config.yaml.example
│   └── config.go                 # viper loader
├── migrations/                   # golang-migrate SQL files
│   └── 0001_init.up.sql / .down.sql
├── pkg/                          # Stateless cross-cutting helpers
│   ├── logger/                   # slog wrapper, request_id ctx helpers
│   ├── response/                 # standard JSON envelope
│   ├── paging/                   # pagination DTO
│   ├── errors/                   # sentinel error registry
│   ├── jwt/                      # token sign/verify
│   ├── hash/                     # bcrypt password hashing
│   ├── idempotency/              # bloom + redis SETNX guard
│   └── validator/                # validator instance
├── internal/
│   ├── server/                   # gin engine, global middleware, router
│   │   ├── server.go
│   │   ├── middleware/           # auth, request_id, ratelimit, recover
│   │   └── router.go             # mounts each domain's port/http
│   ├── dbs/                      # MySQL + Redis + Kafka clients (interfaces)
│   │   ├── mysql.go
│   │   ├── redis.go
│   │   └── kafka.go
│   ├── user/                     # ─ Bounded context: identity & auth
│   │   ├── model/
│   │   ├── dto/
│   │   ├── repository/
│   │   ├── service/
│   │   ├── port/http/
│   │   └── types.go              # exported interfaces (DI ports)
│   ├── event/                    # ─ Events, venues, showtimes
│   │   └── …
│   ├── ticket/                   # ─ Ticket types, prices, seat-map
│   │   └── …
│   ├── inventory/                # ─ Seat / quota holds (Redis Lua)
│   │   └── …
│   ├── order/                    # ─ Checkout, order lifecycle
│   │   └── …
│   ├── payment/                  # ─ Provider gateway, webhook
│   │   └── …
│   ├── promo/                    # ─ Discount codes
│   │   └── …
│   └── notification/             # ─ Email/SMS via Kafka consumer
│       └── …
├── integration_test/             # testcontainers-driven E2E tests
├── internal/mocks/               # generated gomock mocks per domain
├── scripts/                      # k6 load-test scripts, dev helpers
├── deploy/
│   └── docker-compose.yml
├── Makefile
├── go.mod
└── CLAUDE.md
```

### Per-domain folder convention

```
internal/<domain>/
├── types.go                      # exported interfaces (the "ports")
├── model/<domain>.go             # GORM entities (DB shape)
├── dto/<domain>.go               # request/response shapes (HTTP shape)
├── repository/<domain>_repo.go   # unexported impl, returns interface
├── service/<domain>_service.go   # unexported impl, returns interface
└── port/
    └── http/
        ├── handler.go            # gin handlers — depend on interfaces only
        └── route.go              # RegisterRoutes(rg *gin.RouterGroup, h *Handler)
```

Wiring lives **only** in `cmd/server/main.go`. Each layer constructor
returns the interface declared in `internal/<domain>/types.go`. See
`.claude/rules/dependency-injection.md`.

---

## 3. Domains (bounded contexts)

### 3.1 `user`
- Aggregates: `User` (id, email, password_hash, name, phone, role).
- Use cases: register, login (issue JWT pair), refresh, profile, logout.
- Tables: `users`, `refresh_tokens`.

### 3.2 `event`
- Aggregates: `Event` (concert/show), `Venue`, `Showtime` (one event has
  many showtimes; tickets attach to a showtime).
- Use cases: create/update event (admin), list public events, get detail
  (heavily read — cache in Redis with short TTL).
- Tables: `events`, `venues`, `showtimes`.

### 3.3 `ticket`
- Aggregates: `TicketType` (e.g. VIP / Standard, price, total_quota,
  per_user_limit), `Seat` (optional, when seat-map enabled).
- Use cases: define ticket types for a showtime, list available types,
  read seat map.
- Tables: `ticket_types`, `seats`.

### 3.4 `inventory`  ⚡ hot path
- Owns the **single source of truth for "is this still available"**
  during the on-sale window.
- Mechanism:
  1. On showtime go-live, warm Redis with
     `inv:{showtime}:{type}` = remaining count, plus per-seat key
     `seat:{showtime}:{seat_id}` = `available`.
  2. `Hold(userID, items, ttl=10m)` runs a **Lua script** that
     atomically: decrements quota, marks seats `held:{userID}`,
     stores hold record `hold:{holdID}` with TTL.
     Returns hold-id or `ErrSoldOut` / `ErrSeatTaken`.
  3. `Release(holdID)` — Lua script reverses the hold (used on
     cancel / TTL-driven cleanup via keyspace notifications).
  4. `Confirm(holdID)` — flips hold record to `confirmed`, publishes
     `order.reserved` to Kafka (eventual DB persistence).
- Use cases: Hold, Release, Confirm, GetHold.
- Storage: Redis primary; MySQL `holds` table is an audit log written
  asynchronously by a Kafka consumer.

### 3.5 `order`
- Aggregates: `Order` (id, user_id, status, total, idempotency_key),
  `OrderItem`.
- Status FSM: `pending → paid → fulfilled` / `pending → cancelled`
  / `pending → expired`.
- Checkout flow:
  1. Validate user + inputs.
  2. Idempotency guard (bloom → Redis SETNX → DB unique index).
  3. Call `inventory.Hold`.
  4. Apply `promo` (optional).
  5. Create order row in MySQL inside a transaction with
     `SELECT … FOR UPDATE` on the order key.
  6. Return payment intent URL.
- Tables: `orders`, `order_items`, `idempotency_keys`.

### 3.6 `payment`
- Provider-agnostic interface `Gateway` with adapters
  (`stripe`, `midtrans`, mock).
- Webhook handler verifies signature → publishes
  `payment.settled` to Kafka → `order` consumer marks order paid +
  calls `inventory.Confirm`.
- Tables: `payments`, `payment_events` (raw webhook log).

### 3.7 `promo`
- `PromoCode` (code, type=percent/fixed, value, max_uses, expires_at,
  per_user_limit). Atomic decrement of `remaining_uses` via Redis
  with DB reconciliation.
- Tables: `promo_codes`, `promo_redemptions`.

### 3.8 `notification`
- Pure consumer of Kafka topics: `order.paid`, `order.cancelled`,
  `payment.failed`. Renders templates, sends via SMTP / SMS provider.
- No HTTP surface, no DB writes (read-only on user/order for templating).

---

## 4. Cross-cutting: how we survive load

| Risk                  | Mitigation                                                 |
|-----------------------|------------------------------------------------------------|
| Oversell at on-sale   | Redis Lua atomic decrement; MySQL only as audit            |
| Double-submit         | Bloom + Redis SETNX + DB unique on `idempotency_key`       |
| DB hot rows           | Inventory ops never touch MySQL on hot path                |
| Webhook replay        | `payment_events.event_id` UNIQUE                           |
| Money rounding        | All amounts `int64` minor units (`coding-style.md`)        |
| Goroutine leaks       | Every `ctx` cancellable; `errgroup` for fan-out            |
| Slow event detail     | Redis `event:{id}` cache, 30s TTL, jittered                |
| Bot abuse             | IP + user-id token-bucket rate limit (Redis) on `/orders`  |
| Connection storms     | DB pool ≤ 30, Redis pool ≤ 100, Kafka batch produce        |
| Outage of provider    | Circuit breaker around payment gateway calls               |
| Cancel reclaim        | Redis keyspace-notify → release hold; cron sweep fallback  |

---

## 5. Key interfaces (DI ports)

The shapes below are the contract each phase must implement. Concrete
types are unexported; constructors return these interfaces.

```go
// internal/user/types.go
type Repository interface {
    Create(ctx context.Context, u *model.User) error
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    GetByID(ctx context.Context, id string) (*model.User, error)
}
type Service interface {
    Register(ctx context.Context, in dto.RegisterInput) (*dto.AuthOutput, error)
    Login(ctx context.Context, in dto.LoginInput) (*dto.AuthOutput, error)
    Refresh(ctx context.Context, refreshToken string) (*dto.AuthOutput, error)
    Profile(ctx context.Context, userID string) (*dto.UserProfile, error)
}

// internal/event/types.go
type Repository interface {
    Create(ctx context.Context, e *model.Event) error
    Update(ctx context.Context, e *model.Event) error
    GetByID(ctx context.Context, id string) (*model.Event, error)
    List(ctx context.Context, q dto.EventQuery) ([]*model.Event, int64, error)
}
type Service interface {
    Create(ctx context.Context, in dto.CreateEventInput) (*dto.Event, error)
    Detail(ctx context.Context, id string) (*dto.Event, error)
    List(ctx context.Context, q dto.EventQuery) ([]*dto.Event, int64, error)
}

// internal/ticket/types.go
type Repository interface {
    CreateTypes(ctx context.Context, t []*model.TicketType) error
    ListByShowtime(ctx context.Context, showtimeID string) ([]*model.TicketType, error)
}

// internal/inventory/types.go
type Inventory interface {
    Warm(ctx context.Context, showtimeID string, types []dto.QuotaSpec) error
    Hold(ctx context.Context, in dto.HoldInput) (*dto.Hold, error)   // atomic Lua
    Release(ctx context.Context, holdID string) error
    Confirm(ctx context.Context, holdID string) error
    Get(ctx context.Context, holdID string) (*dto.Hold, error)
}

// internal/order/types.go
type Repository interface {
    Create(ctx context.Context, o *model.Order) error
    GetByIdempotencyKey(ctx context.Context, userID, key string) (*model.Order, error)
    UpdateStatus(ctx context.Context, id string, from, to model.OrderStatus) error
}
type Service interface {
    Checkout(ctx context.Context, in dto.CheckoutInput) (*dto.CheckoutOutput, error)
    Get(ctx context.Context, userID, id string) (*dto.Order, error)
    Cancel(ctx context.Context, userID, id string) error
}

// internal/payment/types.go
type Gateway interface {
    CreateIntent(ctx context.Context, in dto.IntentInput) (*dto.Intent, error)
    VerifyWebhook(headers map[string]string, body []byte) (*dto.WebhookEvent, error)
}
type Service interface {
    StartIntent(ctx context.Context, orderID string) (*dto.Intent, error)
    HandleWebhook(ctx context.Context, ev *dto.WebhookEvent) error
}

// internal/promo/types.go
type Service interface {
    Apply(ctx context.Context, code string, userID string, subtotal int64) (int64, error)
    Redeem(ctx context.Context, code, userID, orderID string) error
}
```

---

## 6. HTTP surface (v1)

```
Public
  POST   /api/v1/auth/register
  POST   /api/v1/auth/login
  POST   /api/v1/auth/refresh
  GET    /api/v1/events
  GET    /api/v1/events/:id
  GET    /api/v1/showtimes/:id/tickets
Authenticated (Bearer JWT)
  GET    /api/v1/me
  POST   /api/v1/orders                    # checkout (idempotent)
  GET    /api/v1/orders/:id
  POST   /api/v1/orders/:id/cancel
  POST   /api/v1/holds                     # explicit hold (cart-style)
  DELETE /api/v1/holds/:id
Admin (role=admin)
  POST   /api/v1/admin/events
  PATCH  /api/v1/admin/events/:id
  POST   /api/v1/admin/showtimes/:id/tickets
  POST   /api/v1/admin/promos
Provider webhooks (HMAC verified)
  POST   /webhook/payment/:provider
Ops
  GET    /healthz
  GET    /readyz
  GET    /metrics                          # Prometheus
```

---

## 7. Implementation phases

> Use `/implement-phase <N>` to start a phase. Each phase is self-contained
> and ends with green `golangci-lint run ./...` + `go test -race ./...`
> (see `.claude/rules/pre-commit-checks.md`).

### Phase 1 — Bootstrap
1. `go.mod` (module `github.com/quangdangfit/goticket`, Go 1.25)
2. `Makefile` (build, run, test, lint, mocks, migrate-up, migrate-down)
3. `config/config.go` + `config.yaml.example` (viper)
4. `pkg/logger/logger.go` (slog JSON, request_id ctx key)
5. `pkg/response/response.go` (envelope + error formatter)
6. `pkg/errors/errors.go` (sentinel registry)
7. `pkg/validator/validator.go`
8. `internal/dbs/mysql.go`, `redis.go`, `kafka.go` (interface + impl)
9. `internal/server/middleware/{request_id,recover,ratelimit,auth}.go`
10. `internal/server/server.go` + `router.go` (mounts /healthz, /readyz, /metrics)
11. `cmd/server/main.go` (composes empty server — no domains yet)
12. `deploy/docker-compose.yml` (mysql, redis, kafka, zookeeper)
13. `migrations/0001_init.up.sql` (empty placeholder)

**Exit:** `make run` boots, `curl /healthz` → 200.

### Phase 2 — `user` domain
1. `migrations/0002_users.up.sql` — `users`, `refresh_tokens`
2. `internal/user/{types.go, model/, dto/, repository/, service/, port/http/}`
3. `pkg/jwt`, `pkg/hash`
4. Wire into `cmd/server/main.go`
5. Unit tests for service (mock repo via `internal/mocks/usermock`)
6. Integration test: register → login → /me

### Phase 3 — `event` + `ticket`
1. Migrations: `events`, `venues`, `showtimes`, `ticket_types`, `seats`
2. `internal/event/...`, `internal/ticket/...`
3. Redis read-through cache for `events:{id}` (30s TTL, jitter ±5s)
4. Admin endpoints behind role middleware
5. Unit + integration tests

### Phase 4 — `inventory` (hot path)
1. Lua scripts in `internal/inventory/lua/{hold,release,confirm}.lua`
2. `redisInventory` impl, embedded scripts via `//go:embed`
3. `Warm` populates Redis at showtime publish
4. Integration test: 200 concurrent `Hold` on quota=100 → exactly 100 succeed

### Phase 5 — `order` + idempotency
1. Migrations: `orders`, `order_items`, `idempotency_keys` (UNIQUE
   `(user_id, key)`)
2. `pkg/idempotency` (bloom + Redis SETNX + DB UNIQUE fallback)
3. `orderService.Checkout` orchestrates: idempotency → inventory.Hold →
   create order tx → return intent stub
4. Cancel path → `inventory.Release`
5. Integration test: duplicate checkout returns same order

### Phase 6 — `payment`
1. Migrations: `payments`, `payment_events`
2. `Gateway` interface; `mockGateway` and `stripeGateway` adapters
3. Webhook handler: HMAC verify → publish `payment.settled` to Kafka
4. Order Kafka consumer: marks paid → `inventory.Confirm`
5. Circuit breaker around outbound gateway calls (`sony/gobreaker`)

### Phase 7 — `promo`
1. Migrations: `promo_codes`, `promo_redemptions`
2. Atomic redeem via Redis `DECR` + DB reconcile
3. Hook into `order.Checkout`

### Phase 8 — `notification`
1. Kafka consumer subscribing to `order.paid`, `order.cancelled`
2. SMTP sender (mock in dev)
3. Template registry under `internal/notification/templates/`

### Phase 9 — Observability + hardening
1. OTel tracing middleware; export OTLP
2. Prometheus metrics: `http_requests_total`, `inventory_hold_seconds`,
   `order_checkout_seconds`, `payment_webhook_total`
3. Token-bucket rate limit on `/orders` and `/holds` (Redis)
4. k6 load test under `scripts/k6/checkout.js`

### Phase 10 — Production polish
1. Graceful shutdown (drain in-flight requests, close Kafka, flush logs)
2. Read-replica routing in `dbs/mysql.go` (write vs read pool)
3. Helm chart / k8s manifests under `deploy/k8s/`
4. CI: `.github/workflows/ci.yml` (lint + test + integration)

---

## 8. Conventions recap

- **Money:** `int64` minor units (VND has no decimals).
- **Time:** UTC `time.Time` only.
- **IDs:** ULIDs (`oklog/ulid/v2`) — sortable, no hot AUTO_INCREMENT row.
- **Error wrapping:** `fmt.Errorf("checkout: %w", err)`.
- **Logging:** every log line carries `request_id`, `user_id`,
  `order_id` when in scope.
- **No direct `*gorm.DB` / `*redis.Client` outside repo or dbs packages.**
- **Tests:** table-driven; mocks generated under `internal/mocks/<pkg>/`.
- **Pre-commit gate:** `golangci-lint run ./...` + `go test -race ./internal/... ./pkg/...`.

## Commit Convention

```
feat(payment): add Stripe Checkout Session creation endpoint
fix(webhook): handle out-of-order payment_intent.succeeded after charge.refunded
perf(redis): batch idempotency checks
test(e2e): add full payment flow integration test
```

Types: `feat`, `fix`, `refactor`, `perf`, `test`, `docs`, `chore`, `ci`
Scope: `payment`, `webhook`, `blockchain`, `kafka`, `redis`, `mysql`, `api`, `config`, `bench`

**Author:** commit as configured `git config user.name` / `user.email`. Do **not** append a `Co-Authored-By: Claude …` trailer — repo history has none, keep it that way.
