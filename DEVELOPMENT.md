# Development Guide

## Prerequisites

- Go 1.23+
- PostgreSQL 15+
- golang-migrate CLI
- sqlc
- oapi-codegen
- Docker & Docker Compose (for MinIO)
- Python 3.10+ (for sprite generation scripts)

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Start all services (DB + backend)
make docker-up

# Or start just DB for local development
docker compose -f infrastructure/compose.dev.yml up -d db

# Run migrations
make migrate-up

# Run backend locally
DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable' \
  go run cmd/work-tracker/main.go
```

### Option 2: Local PostgreSQL

```bash
# 1. Start PostgreSQL (manual setup)
# Ensure PostgreSQL is running on localhost:5432

# 2. Create database
createdb workspace

# 3. Run migrations
make migrate-up

# 4. Start server
DATABASE_URL='postgres://postgres:postgres@localhost:5432/workspace?sslmode=disable' \
  go run cmd/work-tracker/main.go
```

## MinIO Setup (Object Storage for Sprites)

### Start MinIO

```bash
# Start MinIO using docker-compose
docker-compose up -d minio

# MinIO API: http://localhost:9000
# MinIO Console (Web UI): http://localhost:9001
# Login: minioadmin / minioadmin
```

### Generate and Upload Placeholder Sprites

```bash
# 1. Install Python dependencies
pip3 install Pillow minio

# 2. Generate placeholder sprite sheets
python3 scripts/generate_placeholder_sprites.py
# Generates 120 sprite files (4 tiers × 10 icons × 3 motions)
# Output: sprites_placeholder/*.png

# 3. Upload sprites to MinIO
python3 scripts/upload_sprites_to_minio.py
# Uploads all sprites to MinIO bucket 'sprites'
```

### Sprite Naming Convention

```
Format: tier{tierNum}-{iconId}_{motion_key}.png
Examples:
  - tier1-01_sleep.png
  - tier1-02_dance.png
  - tier2-05_happy.png

Layout: Horizontal strip (1×N frames, 32px×32px per frame)
Default: 4 frames = 128px × 32px total
```

### Access Sprites

```bash
# Direct URL format
http://localhost:9000/sprites/tier1-01_sleep.png

# Test sprite access
curl -I http://localhost:9000/sprites/tier1-01_sleep.png
```

## API Testing

### Health Check

```bash
curl http://localhost:8000/health
# Expected: ok
```

### Join Command (/in)

```bash
curl -X POST http://localhost:8000/api/commands/join \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "yamada",
    "tier": 1,
    "work_name": "論文執筆"
  }'
```

**Expected Response:**
```json
{
  "session_id": 1,
  "user_id": 1,
  "work_name": "論文執筆",
  "start_time": "2025-10-10T12:00:00Z",
  "planned_end": "2025-10-10T13:00:00Z"
}
```

### Test Duplicate Join (Should return existing session)

```bash
curl -X POST http://localhost:8000/api/commands/join \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "yamada",
    "tier": 1,
    "work_name": "新しい作業"
  }'
```

Should return the same `session_id` if the previous session is still active.

## Database Inspection

```bash
# Connect to database
docker exec -it workspace-dev-db psql -U postgres -d workspace

# Or if using local PostgreSQL
psql -U postgres -d workspace

# SQL queries
SELECT * FROM users;
SELECT * FROM sessions;
```

## Development Commands

```bash
# Show all available commands
make help

# Generate code (sqlc + OpenAPI)
make gen

# Run tests
make test

# Build binary
make build

# Start development environment
make dev-up

# View Docker logs
make docker-logs
```

## Code Generation

### sqlc (Database code generation)

```bash
make gen-sqlc
```

Generates type-safe Go code from SQL queries in `infrastructure/database/query/`.

### OpenAPI (HTTP types generation)

```bash
make gen-openapi
```

Generates HTTP types and Chi server interface from `shared/api/openapi.yaml`.

## Database Migrations

```bash
# Create new migration
make migrate-create NAME=add_new_table

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Project Structure

```
workspace-backend/
├── cmd/work-tracker/     # Entry point
├── domain/               # Domain layer (entities, interfaces)
├── usecase/              # Use case layer (business logic)
├── presentation/         # Presentation layer (HTTP handlers)
├── infrastructure/       # Infrastructure layer (DB, config)
├── shared/               # Shared contracts (OpenAPI)
├── migrations/           # Database migrations
└── scripts/              # Development scripts
```

## Testing Strategy

- **Unit tests**: Domain and usecase layers (with mocks)
- **Integration tests**: Repository layer (with testcontainers)
- **E2E tests**: Full HTTP request/response cycle

Run tests:
```bash
make test
```

## Troubleshooting

### Docker not available in WSL2

Enable Docker Desktop WSL2 integration:
1. Open Docker Desktop settings
2. Navigate to Resources > WSL Integration
3. Enable integration for your WSL2 distro

### Database connection issues

Check DATABASE_URL format:
```
postgres://user:password@host:port/database?sslmode=disable
```

For Docker: `host` should be `db`
For local: `host` should be `localhost`
