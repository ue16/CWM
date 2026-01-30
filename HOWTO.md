# Basic How To

Kanban-style application utilizing cloud technologies, with a Vue.js frontend, Go backend, Postgres/MongoDB database options, and S3-compatible object storage.

## Quick start
1. `docker compose --profile <profile_name> up --build`
	- The profiles are `postgres`, `mongodb-single` and `mongodb-cluster`.
2. Make sure the `DATABASE_URL` environment variable is set correctly in the `backend/.env` file.
3. Open: http://localhost:8080 for the frontend.

## What the project contains
- `frontend/` — Vue.js frontend hosted with Caddy.
- `backend/` — Go backend split into microservices.
- `docker-compose.yml` — services: `frontend`, multiple backend microservices, `database` (Postgres) and `mongodb` (alternative database or cluster), `filestore` (S3-compatible object storage), `proxy` (Caddy on host:8080). Use Docker Compose profiles to switch between single-node MongoDB and a MongoDB cluster.
- `Caddyfile` — reverse-proxy config for frontend and backend services.
- `seed/*` — schema + sample rows; auto-run when a database container initializes.

### Backend
The backend is written in Go and split into multiple microservices:
- Each microservice is responsible for a specific domain (e.g., boards, columns, cards).
- Microservices include their own database connections and CRUD helpers.
- OpenAPI documentation is generated automatically for each microservice.

### Database
The project supports both PostgreSQL and MongoDB databases for persistence. The decision on which database to use is determined by the `DATABASE_URL` environment variable, which can be configured in the `backend/.env` file. By default, the project is set to use MongoDB, but you can switch to PostgreSQL by uncommenting the respective `DATABASE_URL` line in the `.env` file. Additionally, you can use Docker Compose profiles to deploy a MongoDB cluster instead of a single-node MongoDB instance.

- For PostgreSQL: Ensure the `DATABASE_URL` is set to a connection string like `postgres://server:testpass@database:5432/kanban`.
- For MongoDB: Ensure the `DATABASE_URL` is set to a connection string like `mongodb://mongodb:27017/kanban`.

The PostgreSQL database is initialized with a schema and sample data from `seed/database.sql`.

The MongoDB database (single instance) is initialized automatically using the script `seed/kanban.mongodb.js`. This script creates the database, collections, indexes, and seeds it with sample data.

#### MongoDB Cluster

The cluster cannot be set up automatically. Start only the mongodb nodes first:
```
docker compose --profile mongodb-cluster up mongodb-replica-1 mongodb-replica-2 mongodb-replica-3
```

Next, set up the replica set:
```
docker compose --profile mongodb-cluster exec -it mongodb-replica-1 mongosh /initfiles/00-init_cluster.js
```

Then wait a few seconds so the election is settled.
The init script gives the first node a preference as primary, so it should have become the primary node.
Now you can initialize the database:
```
docker compose --profile mongodb-cluster exec -it mongodb-replica-1 mongosh /initfiles/10-kanban.mongodb.js
```


## Environment
- The app expects `DATABASE_URL` (Postgres connection string). Docker Compose sets it for the backend; set it manually if running locally.
- The S3-compatible object storage requires `S3_ENDPOINT`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY` environment variables.
