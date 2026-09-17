# Autonoma

An autonomous AI operations platform for executing real business workflows.

## Phase 0 foundation

Phase 0 provides:

- a minimal Go API with `GET /healthz`
- a separate Python AI runtime skeleton
- Docker-based PostgreSQL and Redis for local development
- formatting, linting, tests, and CI wiring

## Local development

1. Copy `.env.exmaple` to `.env` if you want to overrride defaults.
2. Run `make test`
3. Run `make run-api` and open `http://localhost:8080/healthz`.
4. Run `docker-compose up --build` to start the API and local dependencies.

Docker will store PostgreSQL and Redis data under `./volumes/` in the project root.