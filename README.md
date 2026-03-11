# api-task-user

REST API for task management with role-based access control, built in Go following **Hexagonal Architecture** and **Test-Driven Development**.

---

## Table of Contents

- [Project Summary](#project-summary)
- [Technology Stack](#technology-stack)
- [Software Requirements](#software-requirements)
- [Environment Variables](#environment-variables)
- [Quick Start](#quick-start)
- [Available Make Commands](#available-make-commands)
- [API Endpoints](#api-endpoints)
- [Seed Data & Initial SQL](#seed-data--initial-sql)
- [End-to-End cURL Walkthrough](#end-to-end-curl-walkthrough)
- [Architecture](#architecture)

---

## Project Summary

`api-task-user` is a task management backend that enforces strict **role-based access control** across three user profiles:

| Profile | Capabilities |
|---------|-------------|
| **ADMIN** | Full control over users and tasks. Cannot create other admins. Can only modify or delete tasks in `ASSIGNED` state. |
| **EXECUTOR** | Works on assigned tasks: changes status, and adds comments to expired tasks. |
| **AUDITOR** | Read-only access to all tasks for monitoring and audit purposes. |

Key design decisions:

- **Hexagonal Architecture (Ports & Adapters)** — the domain layer has zero dependencies on HTTP or database frameworks.
- **JWT authentication** — stateless, with the user's profile embedded in the token claims.
- **State machine in the domain** — task transitions (`ASSIGNED → STARTED → FINISHED_*`) are enforced by the `Task` entity itself, not by the database.
- **TDD** — all business logic was written test-first following the Red-Green-Refactor cycle.
- **Expiration rule** — executors cannot change the status of an expired task (`due_date < now()`), but they *can* add a comment to explain what happened.

---

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language | Go | 1.22+ |
| HTTP Router | chi | v5 |
| Database | PostgreSQL | 16 |
| DB Driver | pgx | v5 |
| Authentication | golang-jwt/jwt | v5 |
| Password Hashing | golang.org/x/crypto (bcrypt) | latest |
| Input Validation | go-playground/validator | v10 |
| Migrations | golang-migrate | v4 |
| Mocks (testing) | uber-go/mock (mockgen) | v0.5 |
| Assertions (testing) | testify | v1.10 |
| Config | godotenv | v1.5 |
| UUID | google/uuid | v1.6 |
| Linter | golangci-lint | latest |
| Hot Reload (dev) | air | latest |
| Containers | Docker + Docker Compose | — |

---

## Software Requirements

| Tool | Minimum Version | Install |
|------|----------------|---------|
| Go | 1.22 | https://go.dev/dl |
| Docker | 24.x | https://docs.docker.com/get-docker |
| Docker Compose | v2.x | included with Docker Desktop |
| GNU Make | 3.81+ | pre-installed on macOS/Linux |
| golang-migrate *(optional, local dev)* | v4 | `make tools` |
| golangci-lint *(optional, local dev)* | latest | `make tools` |

---

## Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Runtime environment (`development` / `production`) | `development` |
| `APP_PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `task_manager` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_SSLMODE` | SSL mode for the DB connection | `disable` |
| `DATABASE_URL` | Full connection URL (auto-built from above) | — |
| `JWT_SECRET` | **Required.** Secret key for signing JWT tokens (min. 32 chars) | — |
| `JWT_EXPIRATION_HOURS` | Token lifetime in hours | `24` |
| `BCRYPT_COST` | bcrypt hashing cost factor | `12` |
| `RATE_LIMIT_REQUESTS` | Max requests per time window per IP | `5` |
| `RATE_LIMIT_WINDOW_SECONDS` | Rate-limit time window in seconds | `60` |

> **Security note:** Never commit a real `JWT_SECRET` to version control. Use a randomly generated string of at least 32 characters.

---

## Quick Start

### Option A — Docker Compose (recommended)

```bash
# 1. Copy and configure environment
cp .env.example .env
# Edit .env and set JWT_SECRET

# 2. Build and start the full stack (API + PostgreSQL + migrations)
make docker-up-build

# 3. Verify the API is healthy
curl http://localhost:8080/health
# → HTTP 200
```

### Option B — Local (Go + external PostgreSQL)

```bash
# 1. Copy and configure environment
cp .env.example .env
# Edit .env: set DB_* variables pointing to your local PostgreSQL
# Set JWT_SECRET

# 2. Install development tools
make tools

# 3. Download dependencies
make deps

# 4. Apply database migrations
make migrate-up

# 5. Run the server
make run
# or with hot-reload:
make run-watch
```

---

## Available Make Commands

```
make help             # Show all available commands

# Development
make run              # Start the server
make run-watch        # Start with hot-reload (air)
make build            # Compile binary to ./bin/

# Testing
make test             # Run all tests
make test-unit        # Domain unit tests only
make test-coverage    # HTML coverage report

# Code quality
make lint             # Run golangci-lint
make fmt              # Format code (gofmt + goimports)
make check            # Full quality pipeline (fmt + vet + lint + coverage)

# Database
make db-up            # Start only PostgreSQL
make migrate-up       # Apply pending migrations
make migrate-down     # Revert last migration

# Docker
make docker-up-build  # Rebuild and start stack
make docker-down      # Stop containers
make docker-logs      # Tail API logs

# Mocks
make mocks            # Regenerate testify mocks with mockgen
```

---

## API Endpoints

| Method | Path | Profile required | Description |
|--------|------|-----------------|-------------|
| `GET` | `/health` | Public | Health check |
| `POST` | `/auth/login` | Public | Authenticate and receive JWT |
| `POST` | `/auth/logout` | Any | Invalidate current token |
| `PUT` | `/auth/password` | Any | Change own password |
| `POST` | `/users` | ADMIN | Create a user (EXECUTOR or AUDITOR) |
| `GET` | `/users` | ADMIN | List all users |
| `GET` | `/users/{id}` | ADMIN | Get user by ID |
| `DELETE` | `/users/{id}` | ADMIN | Delete a user |
| `POST` | `/tasks` | ADMIN | Create a task and assign to an executor |
| `DELETE` | `/tasks/{id}` | ADMIN | Delete a task (only if `ASSIGNED`) |
| `GET` | `/tasks/me` | EXECUTOR | List own assigned tasks |
| `PATCH` | `/tasks/{id}/status` | EXECUTOR | Change task status |
| `POST` | `/tasks/{id}/comments` | EXECUTOR | Add comment to an expired task |
| `GET` | `/tasks` | AUDITOR | List all tasks |

---

## Seed Data & Initial SQL

The initial migration (`000001_init.up.sql`) creates the schema and seeds three users. All three have `must_change_password = true`, so **the first action for each user must be to change their password**.

```sql
-- Seed users inserted by the migration
-- Initial temporary password for all three: changeme
-- (bcrypt hash: $2a$12$WvwOq2npSnIB7LxWB3KeBOdOl7itZGnKDvNlH9Mw2OuQl29HILi6.)

SELECT email, profile, must_change_password FROM users;

-- email                        | profile  | must_change_password
-- -----------------------------|----------|---------------------
-- admin@taskmanager.local      | ADMIN    | true
-- executor@taskmanager.local   | EXECUTOR | true
-- auditor@taskmanager.local    | AUDITOR  | true
```

### Useful queries for development

```sql
-- View all users
SELECT id, name, email, profile, must_change_password FROM users;

-- View all tasks with assignee name
SELECT t.id, t.title, t.status, t.due_date, u.name AS assignee
FROM tasks t
JOIN users u ON u.id = t.assignee_id
ORDER BY t.created_at DESC;

-- View comments for a specific task
SELECT c.id, u.name AS author, c.body, c.created_at
FROM comments c
JOIN users u ON u.id = c.author_id
WHERE c.task_id = '<task-uuid>';

-- Manually expire a task (for testing the comment flow)
UPDATE tasks
SET due_date = NOW() - INTERVAL '1 day'
WHERE id = '<task-uuid>';

-- Reset a user's must_change_password flag
UPDATE users SET must_change_password = FALSE WHERE email = 'executor@taskmanager.local';
```

---

## End-to-End cURL Walkthrough

The examples below walk through the full lifecycle: first login → password change → task creation → status transitions → expiration → comment.

> **Base URL:** `http://localhost:8080`
> **Initial password for all seed users:** `changeme`

---

### Step 1 — Admin: first login

```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@taskmanager.local","password":"changeme"}' | jq .
```

**Response:**
```json
{
  "token": "<admin-jwt>",
  "must_change_password": true
}
```

> Save the token: `ADMIN_TOKEN=<admin-jwt>`

---

### Step 2 — Admin: change own password

```bash
curl -s -X PUT http://localhost:8080/auth/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "current_password": "changeme",
    "new_password":     "Admin$1234"
  }'
# → HTTP 204 No Content
```

---

### Step 3 — Executor: first login & password change

```bash
# Login
EXEC_RESP=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"executor@taskmanager.local","password":"changeme"}')

EXEC_TOKEN=$(echo $EXEC_RESP | jq -r '.token')

# Change password
curl -s -X PUT http://localhost:8080/auth/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EXEC_TOKEN" \
  -d '{
    "current_password": "changeme",
    "new_password":     "Executor$1234"
  }'
# → HTTP 204 No Content
```

---

### Step 4 — Auditor: first login & password change

```bash
# Login
AUD_RESP=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"auditor@taskmanager.local","password":"changeme"}')

AUD_TOKEN=$(echo $AUD_RESP | jq -r '.token')

# Change password
curl -s -X PUT http://localhost:8080/auth/password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUD_TOKEN" \
  -d '{
    "current_password": "changeme",
    "new_password":     "Auditor$1234"
  }'
# → HTTP 204 No Content
```

---

### Step 5 — Re-login with new passwords

```bash
# Admin re-login
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@taskmanager.local","password":"Admin1234!"}' \
  | jq -r '.token')

# Executor re-login
EXEC_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"executor@taskmanager.local","password":"Executor1234!"}' \
  | jq -r '.token')

# Auditor re-login
AUD_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"auditor@taskmanager.local","password":"Auditor1234!"}' \
  | jq -r '.token')
```

---

### Step 6 — Admin: get the executor's user ID

The task creation requires the executor's UUID. Retrieve it from the users list:

```bash
curl -s http://localhost:8080/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

**Response (example):**
```json
[
  { "id": "aaaa-...", "name": "System Administrator", "profile": "ADMIN"    },
  { "id": "bbbb-...", "name": "Pedro Executor",        "profile": "EXECUTOR" },
  { "id": "cccc-...", "name": "Juan Auditor",           "profile": "AUDITOR"  }
]
```

```bash
EXECUTOR_ID="bbbb-..."   # replace with the real UUID from the response
```

---

### Step 7 — Admin: create a task

```bash
TASK_RESP=$(curl -s -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"title\":       \"Migrate legacy database\",
    \"description\": \"Export all records from the old MySQL instance and import into PostgreSQL.\",
    \"assignee_id\": \"$EXECUTOR_ID\",
    \"due_date\":    \"2026-12-31T23:59:59Z\"
  }")

echo $TASK_RESP | jq .
TASK_ID=$(echo $TASK_RESP | jq -r '.id')
```

**Response:**
```json
{
  "id":          "<task-uuid>",
  "title":       "Migrate legacy database",
  "description": "Export all records from the old MySQL instance and import into PostgreSQL.",
  "status":      "ASSIGNED",
  "assignee_id": "<executor-uuid>",
  "due_date":    "2026-12-31T23:59:59Z",
  "created_at":  "2026-03-11T10:00:00Z"
}
```

---

### Step 8 — Executor: view own tasks

```bash
curl -s http://localhost:8080/tasks/me \
  -H "Authorization: Bearer $EXEC_TOKEN" | jq .
```

---

### Step 9 — Executor: change task status (ASSIGNED → STARTED)

```bash
curl -s -X PATCH http://localhost:8080/tasks/$TASK_ID/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EXEC_TOKEN" \
  -d '{"status": "STARTED"}'
# → HTTP 204 No Content
```

---

### Step 10 — Auditor: list all tasks

```bash
curl -s http://localhost:8080/tasks \
  -H "Authorization: Bearer $AUD_TOKEN" | jq .
```

---

### Step 11 — Expire the task via SQL

There is no HTTP endpoint to force-expire a task (by design). Use a direct SQL update to simulate expiration for testing purposes:

```bash
# Connect to the database
docker exec -it task-manager-db psql -U postgres -d task_manager

# Inside psql — set due_date to yesterday
UPDATE tasks
SET due_date = NOW() - INTERVAL '1 day'
WHERE id = '<paste-task-uuid-here>';

-- Verify
SELECT id, title, status, due_date FROM tasks WHERE id = '<paste-task-uuid-here>';
\q
```

---

### Step 12 — Executor: try to change status of expired task (expected error)

```bash
curl -s -X PATCH http://localhost:8080/tasks/$TASK_ID/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EXEC_TOKEN" \
  -d '{"status": "FINISHED_SUCCESS"}'
```

**Response (HTTP 400):**
```json
{
  "error": "task is expired — status change not allowed"
}
```

---

### Step 13 — Executor: add a comment to the expired task

Comments are only allowed on expired tasks. This is the mechanism for the executor to document why the task wasn't completed on time.

```bash
curl -s -X POST http://localhost:8080/tasks/$TASK_ID/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EXEC_TOKEN" \
  -d '{
    "body": "The migration was blocked by an unresolved dependency on the legacy auth service. The database team has been notified and the task will be re-scheduled."
  }' | jq .
```

**Response (HTTP 201):**
```json
{
  "id":         "<comment-uuid>",
  "task_id":    "<task-uuid>",
  "author_id":  "<executor-uuid>",
  "body":       "The migration was blocked by an unresolved dependency...",
  "created_at": "2026-03-11T12:30:00Z"
}
```

---

### Step 14 — Admin: logout

```bash
curl -s -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# → HTTP 204 No Content
# The token is now blacklisted and cannot be reused.
```

---

## Architecture

This project is documented using the **4+1 Architectural View Model**. See [`docs/architecture/README.md`](docs/architecture/README.md) for the full reference, which includes:

- **Use Case View** — actors and functional requirements
- **Logical View** — domain model, services, and port contracts
- **Development View** — hexagonal package structure and dependency rules
- **Process View** — sequence diagrams for all major flows
- **Task State Machine** — all valid status transitions with business rules
- **Entity-Relationship** — PostgreSQL physical data model

