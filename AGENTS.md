This is a teaching project in a university master's course called "Cloud, Web & Mobile".

The goal is to build a small Kanban-style application that utilizes cloud technologies.

The frontend is built using Vue.js and resides in the frontend/ folder.
It is hosted using a Caddy container.

The backend is written in Go and resides in the backend/ folder.
It uses Postgres as database and is split into microservices.
It is hosted in custom Docker containers.
We use Swagger/OpenAPI documentation that is generated automatically.

The overall project is started using Docker Compose with the docker-compose.yml file in the root of the project.
CI is used with the GitLab CI available in the remote git.
