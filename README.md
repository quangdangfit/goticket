# GoTicket

High-load event ticketing platform written in Go. Built for flash-sale
traffic (10k+ RPS at on-sale moments) with strict no-oversell, no-double-
charge guarantees.

Architecture: **Domain-Driven Design**. Each bounded context is a folder
under `internal/<domain>/` with the same ports-and-adapters layout
(`types.go` ports, `model/`, `dto/`, `repository/`, `service/`,
`port/http/`). Wiring lives in a single composition root at
`cmd/server/main.go`.

The full design — domain breakdown, key interfaces, HTTP surface, and the
10-phase implementation order — is in [`CLAUDE.md`](./CLAUDE.md).

---

## Stack

| Concern               | Choice                                      |
|-----------------------|---------------------------------------------|
| Language              | Go 1.25                                     |
| HTTP framework        | Gin                                         |
| ORM                   | GORM v2 (MySQL driver)                      |
| Primary DB            | MySQL 8                                     |
| Cache / locks / hold  | Redis 7 (Lua scripts on the hot path)       |
| Async pipeline        | Kafka (`segmentio/kafka-go`)                |
| Migrations            | `golang-migrate`                            |
| Auth                  | JWT (HS256 access + refresh, sha256-stored) |
| Metrics               | Prometheus (`/metrics`)                     |
| Circuit breaker       | `sony/gobreaker` (around payment gateway)   |
| Mocks / tests         | `go.uber.org/mock`, `testcontainers-go`     |
| Logging               | `log/slog` JSON, `request_id`/`user_id` ctx |

---

## Bounded contexts

```
internal/
├── user/          identity & auth (register / login / refresh / me)
├── event/         events, venues, showtimes (Redis read-through cache)
├── ticket/        ticket types, seats, live availability
├── inventory/     ⚡ hot path — Redis Lua atomic hold/release/confirm
├── order/         checkout, FSM, idempotency, SELECT … FOR UPDATE
├── payment/       gateway port + Stripe/mock adapters + webhook
├── promo/         discount codes (percent / fixed) with atomic redeem
└── notification/  Kafka consumer for order.paid / payment.failed
```

Cross-cutting:

```
pkg/
├── logger/        slog wrapper, request-scoped ctx fields
├── response/      standard JSON envelope + error classifier
├── errors/        sentinel registry (ErrNotFound, ErrSoldOut, …)
├── jwt/           HS256 token sign/verify
├── hash/          bcrypt
├── idempotency/   Redis SETNX guard
├── validator/     go-playground/validator instance
└── paging/        page/limit DTO
```

---

## How load is survived

| Risk                  | Mitigation                                                 |
|-----------------------|------------------------------------------------------------|
| Oversell at on-sale   | Redis Lua atomic decrement; MySQL only as audit            |
| Double-submit         | Redis SETNX → DB UNIQUE on `(user_id, idempotency_key)`    |
| DB hot rows           | Inventory ops never touch MySQL on the hot path            |
| Webhook replay        | UNIQUE `(provider, event_id)` on `payment_events`          |
| Money rounding        | All amounts `int64` minor units (no floats)                |
| Slow event detail     | Redis `event:{id}` cache, 30s TTL, ±10% jitter             |
| Bot abuse             | Per-user / per-IP token-bucket rate limit on `/orders`     |
| Provider outage       | `gobreaker` circuit breaker around outbound gateway calls  |
| FSM races             | `SELECT … FOR UPDATE` on order status transitions          |
| Cancel reclaim        | Redis keyspace notifications + cron sweep release          |

---

## HTTP surface (v1)

```
Public
  POST   /api/v1/auth/register
  POST   /api/v1/auth/login
  POST   /api/v1/auth/refresh
  POST   /api/v1/auth/logout
  GET    /api/v1/events
  GET    /api/v1/events/:id
  GET    /api/v1/showtimes/:id/tickets

Authenticated (Bearer JWT)
  GET    /api/v1/me
  POST   /api/v1/orders                  # checkout (idempotent + rate-limited)
  GET    /api/v1/orders/:id
  POST   /api/v1/orders/:id/cancel
  POST   /api/v1/holds                   # cart-style explicit hold
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
  GET    /metrics
```

---

## Quick start

Prerequisites: Go 1.25+, Docker, `golang-migrate`, optional `golangci-lint`.

```bash
# 1. Bring up MySQL + Redis + Kafka
make up

# 2. Apply migrations
DB_DSN="goticket:goticket@tcp(127.0.0.1:3306)/goticket" make migrate-up

# 3. Configure
cp config/config.yaml.example config/config.yaml
# edit secrets if needed (or override with GOTICKET_* env vars)

# 4. Run the API
make run
# → :8080 with /healthz, /readyz, /metrics live

# 5. Smoke test
curl -s localhost:8080/healthz
curl -sX POST localhost:8080/api/v1/auth/register \
  -H 'content-type: application/json' \
  -d '{"email":"a@b.com","password":"password1","name":"Alice"}' | jq
```

### Make targets

| Target              | What it does                                    |
|---------------------|-------------------------------------------------|
| `make build`        | `go build -o bin/server ./cmd/server`           |
| `make run`          | run the server with current config              |
| `make test`         | unit tests (no race)                            |
| `make test-race`    | unit tests with `-race` (the pre-commit gate)   |
| `make test-integration` | spin testcontainers and run E2E suite       |
| `make lint`         | `golangci-lint run ./...`                       |
| `make mocks`        | regenerate `internal/mocks/*` via `mockgen`     |
| `make migrate-up`   | apply pending migrations                        |
| `make migrate-down` | revert one migration                            |
| `make up` / `down`  | docker compose for MySQL / Redis / Kafka        |

### Configuration

`config/config.yaml` is the source of truth. Every key can be overridden
by an env var prefixed `GOTICKET_` (dots → underscores). Example:

```bash
GOTICKET_MYSQL_DSN="user:pw@tcp(prod-db:3306)/goticket?parseTime=true" \
GOTICKET_JWT_ACCESS_SECRET="$(cat /run/secrets/jwt-access)" \
./bin/server
```

---

## Testing

The pre-commit gate (mirrored by CI in `.github/workflows/ci.yml`) is:

```bash
golangci-lint run ./...
go test -race -count=1 -timeout 120s ./internal/... ./pkg/...
```

Unit tests live next to source (`*_test.go`); integration tests under
`integration_test/` use `testcontainers-go` to spin real MySQL, Redis,
and Kafka. The behaviours that *must* be integration tested (because
unit tests can't catch them) are listed in
[`.claude/rules/testing.md`](./.claude/rules/testing.md): `SELECT FOR
UPDATE` under concurrent goroutines, Redis SETNX races, Kafka batch
inserts, and MySQL UNIQUE on `(provider, event_id)`.

### Load testing

```bash
TOKEN=$(curl ... | jq -r .data.access_token)
k6 run -e BASE=http://localhost:8080 \
       -e TOKEN=$TOKEN \
       -e SHOWTIME=<id> \
       -e TICKET_TYPE=<id> \
       scripts/k6/checkout.js
```

The script ramps to 500 vUs and asserts p95 < 500ms with >99% of
responses landing in `{201, 409, 410}`.

---

## Project rules (`.claude/rules/`)

The repo ships with rule files that govern code review and Claude Code
behaviour. They are short and worth a read:

- [`coding-style.md`](./.claude/rules/coding-style.md) — Go style,
  money/time conventions, error wrapping
- [`dependency-injection.md`](./.claude/rules/dependency-injection.md) —
  the cross-package interface rule
- [`security.md`](./.claude/rules/security.md) — secrets, HMAC,
  parameterized SQL, no PII in logs
- [`testing.md`](./.claude/rules/testing.md) — gomock layout, what
  must be integration-tested
- [`pre-commit-checks.md`](./.claude/rules/pre-commit-checks.md) — the
  lint + race gate every commit must pass

Slash commands under `.claude/commands/`:

- `/implement-phase <N>` — execute a numbered phase from `CLAUDE.md`
- `/endpoint <description>` — add a new HTTP endpoint following the
  established handler / service / repo / route layout
- `/review` — payment-system-aware review pass

---

## Status

All 10 implementation phases from `CLAUDE.md` are merged. Outstanding
follow-ups (not blocking):

- `integration_test/` testcontainers suite
- OTel tracing exporter (waiting on collector decision)
- Read-replica routing inside `internal/dbs/mysql.go`
- Helm / k8s manifests under `deploy/k8s/`
- Replace the mock payment provider with the real `stripe-go` SDK

## License

MIT — see [`LICENSE`](./LICENSE).
