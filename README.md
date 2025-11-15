# Interview Stack

A multi-language monorepo that exposes the same Product CRUD API via Fiber (Go), Express (Node.js), and ASP.NET Core Minimal API (.NET 8). Each service can connect to PostgreSQL (pgx/pg/Dapper) or MongoDB (mongo-driver/mongoose/MongoDB.Driver) to persist the same entity, includes structured logging, health checks, and Docker builds. ⚠️ This is not production ready: credentials live in the sample config files and the stack is not wired to source secrets from a secure store.

## Repository Layout

```
/interview-stack
├── go-service/          # Fiber service (config, db, handlers, routes)
├── node-service/        # Express service (config, db, handlers, routes)
├── dotnet-service/      # .NET 8 Minimal API
├── migrations/          # PostgreSQL schema bootstrap
├── docker-compose.yml   # Runs databases + all services
└── README.md
```

Each service folder ships with a `.env.example`, a `wait-for.sh` helper, and its own Dockerfile.

## Prerequisites

- Docker + Docker Compose v2
- Ports `8081-8083`, `5432`, `27017` available on your machine

## Quick Start

```bash
docker compose up --build
```

Compose will:

- Build the three services
- Start PostgreSQL (with migrations mounted from `migrations/001_init.sql`)
- Start MongoDB with persistent volumes
- Wait for both databases before launching each API
- Start Prometheus on `http://localhost:9090`
- Start Grafana on `http://localhost:3000` with the bundled dashboard/datasource

Once running you can hit:

- Go service: `http://localhost:8081`
- Node service: `http://localhost:8082`
- .NET service: `http://localhost:8083`

### Edge Routing

Nginx fronts every API on `http://localhost:8080` (see `nginx/nginx.conf`). Each path segment rewrites to the matching service so you can exercise the stack through a single origin, e.g. `http://localhost:8080/go/products`, `/node/products/{id}`, or `/dotnet/health`. The router strips the prefix before proxying, adds JSON-friendly headers, and is the recommended URL surface when demoing or benchmarking the platform.

### API Docs (Swagger/OpenAPI)

Every runtime ships its own OpenAPI definition and Swagger UI so you can compare implementations or import the spec into clients:

- **Go** (Fiber Swagger): `http://localhost:8081/swagger/index.html` (or via Nginx `http://localhost:8080/go/swagger/index.html`). Raw spec: `http://localhost:8081/swagger.json`.
- **Node** (swagger-ui-express): `http://localhost:8082/swagger` (`http://localhost:8080/node/swagger`). Raw spec: `http://localhost:8082/swagger.json`.
- **.NET** (Swashbuckle): `http://localhost:8083/swagger` (`http://localhost:8080/dotnet/swagger`). Raw spec: `http://localhost:8083/swagger/v1/swagger.json`.

Use the docs to trigger CRUD operations visually or export the schema for contract tests.

### Shared Endpoints

All services implement:

- `GET    /health`
- `GET    /products`
- `GET    /products/{id}`
- `POST   /products`
- `PUT    /products/{id}`
- `DELETE /products/{id}`

Request body for `POST/PUT`:

```json
{
  "name": "Laptop",
  "price": 1999.99,
  "tags": ["hardware", "premium"]
}
```

Responses mirror the schema:

```json
{
  "id": "uuid-string",
  "name": "Laptop",
  "price": 1999.99,
  "tags": ["hardware", "premium"],
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Environment Variables

Each service loads settings from the environment with sane defaults (see `*.env.example`). The critical values are:

- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `MONGO_URI`, `MONGO_DATABASE`, `MONGO_COLLECTION`
- `<SERVICE>_PORT` (8081/8082/8083)

When running via Compose these are injected automatically.

### Selecting a Store

Services now target a single backing store at a time instead of dual-writing to PostgreSQL and MongoDB. Pick the store by setting `database.type` to `sql` or `mongo` in the service configuration:

- Go: `go-service/config.yaml`
- Node: `node-service/config.yaml`
- .NET: `dotnet-service/src/appsettings*.json`

The Docker Compose setup keeps the default `sql` option for all services, but you can bake new config files into the containers or override them via mounted volumes to switch to MongoDB.

### Cache, SWR & Write-Behind

- **Cache/SWR**: Redis (`redis:6379`) backs a Stale-While-Revalidate layer configured via each service's `cache.*` section. Reads first attempt to serve fresh data (`defaultTTL`, 30s by default). Once the entry ages past `defaultTTL` but before `staleTTL` (60s), requests still return the cached payload while the service refreshes it in the background to minimize latency spikes. Cache invalidation happens on every write and after the write-behind worker replays events.
- **Write-Behind**: Setting `writeBehind.enabled=true` publishes mutations to the Redis Stream `products_write_queue`. Requests can respond after enqueuing, offloading the heavy writes to the background worker (`WriteBehindWorker` / `RedisWriteQueueProducer`) which batches operations based on `writeBehind.batchSize`/`flushInterval`. Disable it to force fully synchronous writes for stricter consistency.
- **Fallbacks**: If Redis is unavailable at startup, the cache/write-behind features automatically disable themselves, so you can still run the APIs locally.

### C4 Diagrams

Detailed C4 views live under `diagrams/` and embed Mermaid source so you can paste them into any renderer:

- [Contexto](diagrams/structurizr-contexto.md)
- [Contenedores](diagrams/structurizr-contenedores.md)
- [Componentes Go](diagrams/structurizr-componentes_go.md)
- [Flujo POST /products](diagrams/structurizr-post_products.md)

### SQL Schema

PostgreSQL migrations live under `migrations/`. The mounted script creates the `products` table, UUID extension, and update trigger. Apply manually when running outside Docker:

```bash
psql postgres://postgres:postgres@localhost:5432/productsdb -f migrations/001_init.sql
```

## Service Notes

- **Go (Fiber)**: leverages `pgxpool` + official Mongo driver, wraps CRUD logic in `ProductRepository`, and emits structured logs via `slog`.
- **Node (Express)**: uses `pg` pooling combined with `mongoose` for MongoDB, `pino` for logging, and consistent validation in the handler layer.
- **.NET 8 Minimal API**: `Npgsql` + `Dapper` for SQL, `MongoDB.Driver` for NoSQL, clean DI wiring, and strongly typed handlers.

### Architecture Principles

The monorepo is structured around SOLID principles:

- **Single Responsibility**: controllers just translate HTTP ↔ DTOs, use cases own business rules, repositories abstract persistence, and cache/queue layers encapsulate infra.
- **Open/Closed**: configuration toggles (database type, cache, write-behind, metrics) let you extend behaviour without touching code paths.
- **Liskov Substitution / Interface Segregation**: every ProductStore implementation (Postgres/Mongo) honors the same interface so use cases don't change when swapping stores.
- **Dependency Inversion**: constructors receive interfaces (loggers, repositories, queues, clocks) via dependency injection, keeping business logic decoupled from frameworks.

## Observability

Every service now exposes `/metrics` (configurable via each service's `metrics` section) using Prometheus client libraries:

- `http_requests_total` & `http_request_duration_seconds` labeled by service/method/route/status
- Cache counters `cache_hits_total` / `cache_misses_total`
- Write-behind batch gauges/counters: `write_behind_lag_seconds`, `write_behind_queue_length`, `write_behind_batch_size`, `write_behind_batch_duration_seconds`, `write_behind_errors_total`

The top-level `observability/prometheus.yml` config scrapes the three APIs. Grafana automatically provisions:

- a Prometheus datasource pointing at `http://prometheus:9090`
- the sample dashboard stored at `observability/grafana-dashboard.json`

Log in to Grafana at `http://localhost:3000/d/interview-observability/interview-stack-observability` (alias of the default `http://localhost:3000`) with `admin/admin`, then switch the `service` dropdown to view per-service request rates, latency, cache hit ratios, and write-behind queue health.

## Development Tips

- Run services individually by exporting the env vars from the `.env.example` of the respective service and executing the platform runtime (`go run ./cmd/server`, `node src/server.js`, or `dotnet run`).
- Logs are JSON-friendly across services; adjust `LOG_LEVEL` (Node) or change handlers as needed.
- Update the schema for the store you selected (SQL or Mongo). Since the APIs no longer mirror writes to both stores, switch types only after you have applied the equivalent migrations in each backend.

Enjoy building your interview stack!
