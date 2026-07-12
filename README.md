# PR Reviewer Assignment Service

A REST API service for managing teams and automatically assigning pull request reviewers based on team membership.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green)

## Tech Stack

- **Go** (1.26) — HTTP server, business logic
- **PostgreSQL** (16) — data persistence
- **Docker** & **Docker Compose** — containerized development and deployment
- **golang-migrate** — database schema migrations
- **pgx/v5** — PostgreSQL driver

## Architecture

The project follows **clean architecture** with strict layer separation:

```
cmd/server/main.go          → entry point, dependency wiring
internal/
├── config/                  → environment configuration
├── middleware/              → HTTP middleware (logging, recovery)
├── model/                   → domain entities and error mapping
├── repository/              → data access layer (PostgreSQL)
├── service/                 → business logic and validation
└── handler/                 → HTTP handlers and request/response formatting
```

Each layer depends only on abstractions (interfaces), following the **Dependency Inversion Principle**:
- `service` defines `Repository` interfaces it needs
- `handler` defines `Service` interfaces it needs
- Concrete implementations are wired in `main.go`

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/team/add` | Create a team with members |
| `GET` | `/team/get` | Get a team and its members |
| `POST` | `/users/setIsActive` | Activate/deactivate a user |
| `POST` | `/pullRequest/create` | Create PR with automatic reviewer assignment |
| `POST` | `/pullRequest/merge` | Merge a pull request (idempotent) |
| `POST` | `/pullRequest/reassign` | Reassign a reviewer |
| `GET` | `/users/getReview` | Get all PRs assigned to a user |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Git

### 1. Clone and configure

```bash
git clone https://github.com/kihcnxlehp/pr-reviewer-service.git
cd pr-reviewer-service

# Create environment file from example
cp .env.example .env
```

### 2. Run

```bash
docker-compose up --build
```

The service will:
1. Start PostgreSQL
2. Apply database migrations automatically
3. Start the HTTP server on port 8080

### 3. Test

```bash
# Health check
curl http://localhost:8080/health

# Create a team
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true}
    ]
  }'

# Get team
curl http://localhost:8080/team/get?team_name=backend

# Deactivate a user
curl -X POST http://localhost:8080/users/setIsActive \
  -H "Content-Type: application/json" \
  -d '{"user_id": "u1", "is_active": false}'
```

## Database Schema

```sql
teams (team_name PK)
  └── users (user_id PK, team_name FK, username, is_active)
        └── pull_requests (pull_request_id PK, author_id FK, status)
              └── pr_reviewers (pull_request_id FK, user_id FK)
```

Migrations are stored in `migrations/` and applied automatically on startup.

## Key Design Decisions

- **Transactional operations**: team creation and PR creation use database transactions to ensure atomicity
- **Optimistic concurrency**: unique constraint violations are caught and mapped to domain errors (`TEAM_EXISTS`, `PR_EXISTS`)
- **Strict input validation**: all endpoints validate required fields before hitting the database
- **Safe error responses**: internal errors (DB failures) are hidden from clients; only domain errors are exposed
- **Request size limiting**: 1 MB max body size to protect against DoS
- **Graceful shutdown**: the server waits for active requests to complete before stopping
- **Structured by domain**: each entity (team, user, pull_request) has its own files across all layers

## Project Structure

```
.
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/config.go         # Configuration from env vars
│   ├── model/
│   │   ├── model.go             # Domain entities
│   │   └── errors.go            # Error codes and HTTP status mapping
│   ├── middleware/
│   │   ├── logging.go           # Request logging with time
│   │   └── recovery.go          # Panic recovery
│   ├── repository/
│   │   ├── team.go              # Team data access
│   │   ├── user.go              # User data access
│   │   └── pull_request.go      # PR data access
│   ├── service/
│   │   ├── team.go              # Team business logic
│   │   ├── user.go              # User business logic
│   │   └── pull_request.go      # PR business logic
│   └── handler/
│       ├── handler.go           # Shared HTTP helpers and routing
│       ├── team.go              # Team endpoints
│       ├── user.go              # User endpoints
│       └── pull_request.go      # PR endpoints
├── migrations/                   # SQL migration files
├── .env.example                  # Environment variable template
├── docker-compose.yml            # Multi-container setup
├── Dockerfile                    # Multi-stage build
└── go.mod                        # Go module definition
```

## License

MIT