# Basic How To

Kanban-style application utilizing cloud technologies, with a Vue.js frontend, Go backend, Postgres database, and S3-compatible object storage.

## Quick start
1. `docker compose up --build`
2. Open: http://localhost:8080 for the frontend.

## What the project contains
- `frontend/` — Vue.js frontend hosted with Caddy.
- `backend/` — Go backend split into microservices.
- `docker-compose.yml` — services: `frontend`, multiple backend microservices, `database` (Postgres) and `mongodb` (alternative database), `filestore` (S3-compatible object storage), `proxy` (Caddy on host:8080).
- `Caddyfile` — reverse-proxy config for frontend and backend services.
- `seed/*` — schema + sample rows; auto-run when a database container initializes.

### Backend
The backend is written in Go and split into multiple microservices:
- Each microservice is responsible for a specific domain (e.g., boards, columns, cards).
- Microservices include their own database connections and CRUD helpers.
- OpenAPI documentation is generated automatically for each microservice.

### Database
The project supports both PostgreSQL and MongoDB databases for persistence. The decision on which database to use is determined by the `DATABASE_URL` environment variable, which can be configured in the `backend/.env` file. By default, the project is set to use MongoDB, but you can switch to PostgreSQL by uncommenting the respective `DATABASE_URL` line in the `.env` file.

- For PostgreSQL: Ensure the `DATABASE_URL` is set to a connection string like `postgres://server:testpass@database:5432/kanban`.
- For MongoDB: Ensure the `DATABASE_URL` is set to a connection string like `mongodb://mongodb:27017/kanban`.

The PostgreSQL database is initialized with a schema and sample data from `seed/database.sql`.

The MongoDB database is initialized automatically using the script `seed/kanban.mongodb.js`. This script creates the database, collections, indexes, and seeds it with sample data.

## Environment
- The app expects `DATABASE_URL` (Postgres connection string). Docker Compose sets it for the backend; set it manually if running locally.
- The S3-compatible object storage requires `S3_ENDPOINT`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY` environment variables.
