# Basic How To

Minimal Go service demonstrating a simple Kanban schema (boards, columns, cards).

## Quick start
1. `docker compose up --build`
2. Open: http://localhost:8080

## What the project contains
- `Dockerfile` — builds the Go binary (exposes port 80).
- `docker-compose.yml` — services: `web` (app), `database` (Postgres, seeds DB), `proxy` (Caddy on host:8080).
- `Caddyfile` — reverse-proxy to `web:80`.
- `go.mod` — module and dependencies.
- `main.go` — entrypoint: runs DB demo then starts the HTTP server.
- `internal/database.go` — DB connection (`NewDB`) and CRUD helpers for boards/columns/cards.
- `internal/database_model.go` — `Board`, `Column`, `Card` structs with `String()` helpers.
- `internal/server.go` — small HTTP server that serves the demo output at `/`.
- `seed/database.sql` — schema + sample rows; auto-run when Postgres container initializes.

## Environment
- The app expects `DATABASE_URL` (Postgres connection string). Compose sets it for `web`; set it manually if running locally.

## Notes
- `main.go` creates sample records on each run — the DB will grow if you rerun without resetting.
- On the first run, the app will not be able to access the DB as Postgres is still initializing it. Simply stop compose after a few seconds and start it again.
