# Architecture

## Project structure
```bash
├── api/             # OpenAPI spec for the backend
├── cmd/
│   ├── api          # Entry point for the backend app
│   └── seed         # Entry point for the db seed app, useful during development
├── deploy/          # All stuff related to deployment of this app
├── docs/            # You're here, and reading the docs :D
├── e2e/             # E2E tests for the API (go)
├── go.mod
├── go.sum
├── infra/           # Configurations for infrastructure(grafana, loki, prometheus, caddy)
├── internal/        # The core application (go)
│   ├── config       # > app config duh
│   ├── dtos         # > data transfer objects
│   ├── events       # > message publishing
│   ├── hasher       # > general interface to work with hash
│   ├── jwtutil      # > jwt token helpers
│   ├── logger       # > log/slog configuration
│   ├── metrics      # > Prometheus metrics
│   ├── models       # > domain entities
│   ├── oauth        # > interface to work with OAuth2 providers
│   ├── service      # > business logic (auth, notes, users
│   ├── store        # > persistence layer(postgres, redis)
│   └── transport    # > transport layer(http handlers)
├── mailer/          # mailer service (go)
├── migrations/      # DB migrations live here (sql)
├── Taskfile.yml     # task file with all tasks for the app development
└── web/             # The frontend app (elm)
    ├── review       # > elm-review configuration
    ├── static       # > all static file(images, fonts, styles, etc)
    ├── src
    └── tests
```

## Data flow
1. Frontend/Api -> Backend
  - The Elm frontend ([web/](/web)) sends requests to the backend API.
2. Transport layer (http handlers)
  - User provided data gets mapped into DTO and passed to service layer.
  - DTOs are only used between transport and services; the rest of the application uses domain models.
3. Business logic (services)
  - Services receive DTOs, map them into domain models, and implement the use case (auth, notes, users).
  - Services may interact with:
    - Persistence (store → Postgres, Redis)
    - Events (events → NATS)
    - Utilities (`jwtutil`, `hasher`)
4. Persistence
  - Postgres stores durable data (users, notes).
  - Redis handles caching/ephemeral state.
5. Background tasks (fire-and-forget)
  - Some operations, like sending verification emails, are published as NATS events.
6. Response
  - After business logic is complete, domain models are mapped back to DTOs, and returned via transport layer.
  - Elm frontend updates the UI accordingly.
7. Observability
  - Services log via logger and export Prometheus metrics (`metrics`).
  - Grafana, Loki, and Prometheus provide monitoring.

## Docker
The project uses multiple Dockerfiles with a multi-stage build approach:
- `builder.Dockerfile` – builds Go binaries and caches dependencies.
- `runtime.Dockerfile` – lightweight Alpine base image with system packages preinstalled (used by app images to speed up builds).
- `core.Dockerfile` – packages the main backend API service.
- `mailer.Dockerfile` – packages the mailer service.

## Services
- Core service (go)
  - Exposes RestAPI (see. [API](/docs/API.md))
  - Handles authentication
  - Manages notes life cycle
- Mailer
  - Listens for events from the core.
  - Sends account confirmation and password reset emails.
  - Provides its own Prometheus metrics.
  - **NOTE:** all events of the service is documented [here](/mailer/)

## Frontend
- Built with [elm.land](https://elm.land)
- Uses elm-review rules for consistency and quality.
- Compiled into a static bundle (web/dist), served by Caddy.
