# Repository Guidelines

## Project Structure & Module Organization
- `cmd/work-tracker/main.go`: HTTP/WebSocket backend entrypoint; keep configs and flags here minimal.
- `domain/` houses entities and interfaces; `usecase/command|query|session/` contains business logic; keep I/O-free.
- `infrastructure/` provides adapters (config, logger, database); SQL queries live in `infrastructure/database/query/`, generated code in `infrastructure/database/sqlc/`.
- `presentation/http/` serves the OpenAPI surface; DTOs are generated; handlers orchestrate usecases. `presentation/ws/` streams state to the frontend overlay.
- `shared/api/openapi.yaml` defines contracts; `scripts/` and `Makefile` automate generation, migrations, and local tooling. `migrations/` holds golang-migrate files.

## Build, Test, and Development Commands
- `make dev-up`: start dev DB and apply migrations (Docker based).
- `make docker-up | docker-down | docker-logs`: full compose lifecycle for backend + DB.
- `make migrate-up` / `make migrate-down` / `make migrate-create NAME=add_table`: run or scaffold migrations.
- `make gen` (or `make gen-sqlc`, `make gen-openapi`): regenerate sqlc and OpenAPI types/servers/clients.
- `DATABASE_URL=postgres://.../workspace?sslmode=disable make run` or `go run cmd/work-tracker/main.go`: run locally.
- `make build`: produce `bin/work-tracker` binary.

## Coding Style & Naming Conventions
- Go 1.23; format with `gofmt`/`goimports` before pushing.
- Comments are Japanese; auto-generated files (`infrastructure/database/sqlc/`, `presentation/http/dto/`) are exempt.
- Use `usecase` (not `useCase`) and `repository` (not `repo`) in identifiers per `CODING_GUIDELINES.md`.
- Keep handler/request DTOs thin; route logic to usecases; avoid cross-layer imports.

## Testing Guidelines
- `make test`: full suite (unit + integration) with race + cover; expects `DATABASE_URL=...workspace_test`.
- `make test-unit`: fast checks without DB; `make test-integration`: DB-backed cases.
- Test function names in English; `t.Run` descriptions in Japanese describing scenario/expectation.
- For integration tests, seed via migrations only; avoid test-specific SQL in code.

## Commit & Pull Request Guidelines
- Follow existing history: `feat: ...`, `chore: ...`, `fix: ...` with concise Japanese descriptions (include command name/scope when relevant).
- PRs should describe behavior changes, DB/migration impacts, and commands run (`make test`, `make gen` if applicable); link issues or tickets.
- Include request/response samples or screenshots when UI/API behavior changes; note env vars required to reproduce.

## Security & Config Notes
- Provide `DATABASE_URL` via env for app and tests; do not commit secrets. Prefer `.env.local` patterns ignored by git.
- Run `make migrate-up` after pulling schema changes; keep migration filenames sequential and descriptive.
- Docker default ports defined in `infrastructure/compose.dev.yml`; avoid conflicting host ports when running locally.
