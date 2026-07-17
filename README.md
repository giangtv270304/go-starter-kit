# go_starter_kit

A minimal Go service starter kit built on [gframework](https://github.com/andyle182810/gframework), Echo v5, PostgreSQL, and Valkey (Redis-compatible). It ships with a working `/health` endpoint, Swagger docs generation, structured logging, graceful shutdown, and linting — ready to extend with real handlers and repositories.

## Tech stack

- **Go 1.26**
- **Echo v5** — HTTP routing (via `gframework/httpserver`)
- **gframework** — opinionated wrappers for HTTP server, metrics server, service lifecycle (`runner`), Postgres pool, Valkey client
- **PostgreSQL** (pgx v5) + [golang-migrate](https://github.com/golang-migrate/migrate) for schema migrations
- **Valkey** (Redis-compatible) for caching/queues
- **resty** — outbound HTTP client
- **swaggo/swag** + **echo-swagger** — OpenAPI docs generation
- **golangci-lint**, **gofumpt**, **gci** — linting and formatting

## Project structure

```
.
├── main.go                        # entrypoint: wires config, infra clients, HTTP/metric servers, runner
├── internal/
│   ├── config/config.go           # env-based configuration (caarlos0/env)
│   ├── repo/repo.go                # data access layer (Postgres)
│   └── service/
│       ├── service.go              # service struct + constructor, holds repo/resty/valkey/config
│       └── check_health.go         # example handler: GET /health
├── apispec/                        # generated Swagger docs (swagger.yaml, docs.go) — do not edit by hand
├── db/migrations/                  # SQL migration files for golang-migrate (create with `make migrate-create`)
├── docker-compose.yml               # local Postgres + Valkey + Terraform DB setup
├── Dockerfile                       # multi-stage build (Alpine)
├── .env.example                     # template for local environment variables
├── Makefile                         # build, lint, migration, and API doc commands
└── .golangci.yaml                   # linter configuration
```

## Using this as a starter kit for a new project

To clone this repo and rebrand it for a new project (e.g. `my_new_service`):

1. Clone with the new folder name and drop the old git history:
   ```
   git clone git@github.com:TranVuGiang/go-starter-kit.git my_new_service
   cd my_new_service
   rm -rf .git && git init
   ```
2. Update the Go module path in `go.mod` (first line):
   ```
   module github.com/<your-username>/my_new_service
   ```
3. Rewrite all imports and the service name to match:
   ```
   grep -rl "github.com/go_starter_kit" --include="*.go" . | xargs sed -i '' 's|github.com/go_starter_kit|github.com/<your-username>/my_new_service|g'
   ```
   Then update `serviceName = "go_starter_kit"` in `main.go`.
4. Replace remaining references to `go_starter_kit` in config/CI files:
   - `Makefile` line 1 — `SERVICE_NAME = go_starter_kit` (controls the `make build` output binary name)
   - `Dockerfile` lines 42/59 — `COPY --from=builder /src/go_starter_kit` and `ENTRYPOINT ["/app/go_starter_kit"]` — **must match `SERVICE_NAME`** above, or the Docker build will fail to find the binary
   - `docker-compose.yml` — container names and `TF_VAR_db_*`/`TF_VAR_schema_name`
   - `.github/workflows/prod_pipeline.yml` — `IMAGE_NAME`
5. Verify everything builds:
   ```
   go mod tidy
   go build ./...
   ```
6. Point the repo at its new remote (if pushing to a new GitHub repo):
   ```
   git remote add origin git@github.com:<your-username>/<new-repo-name>.git
   ```

## Prerequisites

- Go 1.26+
- Docker & Docker Compose (for local Postgres/Valkey)
- [`golang-migrate` CLI](https://github.com/golang-migrate/migrate) — only needed for `make migrate-*` targets
- **macOS only**: Xcode Command Line Tools license accepted (`sudo xcodebuild -license`) — required before `go install`-based tools (`swag`, `gofumpt`, `golangci-lint`, `gci`) can fetch/build

## Getting started

1. Copy the env template and fill in secrets:
   ```
   cp .env.example .env
   ```
2. Start local infrastructure:
   ```
   docker compose up -d postgres valkey
   ```
3. (Optional) Run DB migrations if you have migration files under `db/migrations/`:
   ```
   make migrate-up-local
   ```
4. Run the service:
   ```
   go run .
   ```
5. Check it's alive:
   ```
   curl localhost:8081/health
   ```

## Configuration

All configuration is loaded from environment variables (see `internal/config/config.go` for defaults). Key groups:

| Group         | Prefix                                                                                               | Notes                                         |
| ------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| Common        | `GRACEFUL_SHUTDOWN_PERIOD`                                                                           | Shutdown grace period for all services        |
| Logging       | `LOG_LEVEL`                                                                                          | e.g. `debug`, `info`                          |
| Metric server | `METRIC_SERVER_*`                                                                                    | Exposes `/status` and `/metrics` (Prometheus) |
| HTTP server   | `HTTP_SERVER_*`, `HTTP_ENABLE_CORS`, `HTTP_ALLOW_ORIGINS`, `HTTP_BODY_LIMIT`, `HTTP_SKIP_REQUEST_ID` | Main API server                               |
| Swagger       | `SWAGGER_HOST`, `SWAGGER_ENABLED`                                                                    | Enables `/swagger/*` when `true`              |
| Postgres      | `POSTGRES_*`                                                                                         | Connection pool + migration source            |
| Migration     | `MIGRATION_ENABLED`, `MIGRATION_SOURCE`                                                              | Auto-run migrations on startup                |
| Valkey        | `VALKEY_*`                                                                                           | Connection pool for cache/queues              |

`.env` is gitignored — never commit real secrets. Use `.env.example` as the source of truth for available keys.

## Makefile commands

| Command                                | Description                                                                         |
| -------------------------------------- | ----------------------------------------------------------------------------------- |
| `make build`                           | Cross-compile the binary (`go_starter_kit`)                                         |
| `make migrate-create MIGRATION=<name>` | Scaffold a new migration file pair in `db/migrations`                               |
| `make migrate-up-local`                | Apply all pending migrations against the local Postgres (from `docker-compose.yml`) |
| `make migrate-down-local`              | Roll back the last migration                                                        |
| `make lint`                            | Run `go mod tidy`, `gofumpt`, `go vet`, and `golangci-lint`                         |
| `make gci`                             | Fix import ordering                                                                 |
| `make api-doc`                         | Regenerate Swagger docs into `apispec/` via `swag init`                             |
| `make view-api-doc`                    | Serve the generated spec locally with Swagger UI (Docker)                           |

## API docs

Swagger annotations live next to their handlers (see `internal/service/check_health.go`). After changing/adding annotated handlers, regenerate docs with:

```
make api-doc
```

With `SWAGGER_ENABLED=true`, the UI is served at `/swagger/index.html`.

## Adding a new endpoint

1. Add a handler method + request/response types in `internal/service/` (follow the pattern in `check_health.go`).
2. Register the route in `main.go` inside `(*application).registerRoutes`.
3. If it needs data access, add methods to `internal/repo/repo.go` (or a new repo file) using the injected `postgres.DBPool`.
4. Run `make api-doc` if you added Swagger annotations, then `make lint` before committing.

## Docker

```
docker compose up -d          # postgres, valkey, and Terraform-based DB bootstrap
make build                    # produces the ./go_starter_kit binary
docker build -t go_starter_kit .
```

The Docker image copies `db/migrations` into the final image so `MIGRATION_ENABLED=true` can run migrations at container startup.
