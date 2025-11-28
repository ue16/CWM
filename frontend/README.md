# kanban-frontend

## Project Setup

```sh
yarn
```

### Compile and Hot-Reload for Development

The VITE_BACKEND_BASE_URL tells vite where to find the backend,
since it is on another route when no reverse proxy is used.

```sh
VITE_BACKEND_BASE_URL="http://localhost:8080" yarn dev
```

With NodeJS 25:
```sh
VITE_BACKEND_BASE_URL="http://localhost:8080" NODE_OPTIONS="--no-webstorage" yarn dev
```

### Type-Check, Compile and Minify for Production

```sh
yarn build
```
