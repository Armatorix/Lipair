.PHONY: help up down build logs restart \
        backend-build backend-run backend-test backend-lint backend-tidy \
        frontend-install frontend-dev frontend-build frontend-lint \
        db-migrate db-reset generate-api

# ── Default target ───────────────────────────────────────────────────────────
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}' | sort

# ── Docker Compose ───────────────────────────────────────────────────────────
up: ## Start all services (builds if needed)
	docker compose up --build -d

up-logs: ## Start all services and follow logs
	docker compose up --build

down: ## Stop and remove containers
	docker compose down

build: ## Build all Docker images
	docker compose build

logs: ## Follow logs from all services
	docker compose logs -f

logs-backend: ## Follow backend logs
	docker compose logs -f backend

logs-frontend: ## Follow frontend logs
	docker compose logs -f frontend

restart: ## Restart all services
	docker compose restart

restart-backend: ## Restart only the backend service
	docker compose restart backend

ps: ## Show running service status
	docker compose ps

# ── Backend (Go) ─────────────────────────────────────────────────────────────
backend-build: ## Build the Go binary locally
	cd backend && go build -o bin/server ./cmd/server

backend-run: ## Run the backend locally (requires postgres running)
	cd backend && go run ./cmd/server

backend-test: ## Run backend tests
	cd backend && go test ./...

backend-lint: ## Run Go linter (requires golangci-lint)
	cd backend && golangci-lint run ./...

backend-vet: ## Run go vet
	cd backend && go vet ./...

backend-tidy: ## Tidy and verify Go modules
	cd backend && go mod tidy && go mod verify

# ── Frontend (Node) ───────────────────────────────────────────────────────────
frontend-install: ## Install frontend dependencies
	cd frontend && npm install

frontend-dev: ## Start frontend dev server (proxies /api to localhost:8080)
	cd frontend && npm run dev

frontend-build: ## Build frontend for production
	cd frontend && npm run build

frontend-lint: ## Lint frontend TypeScript/React code
	cd frontend && npm run lint

frontend-preview: ## Preview the production build locally
	cd frontend && npm run preview

# ── Database ─────────────────────────────────────────────────────────────────
db-shell: ## Open a psql shell to the local database
	docker compose exec postgres psql -U chessmgr -d chessmgr

db-reset: ## Drop and recreate the local database (WARNING: destroys all data)
	docker compose down -v
	docker compose up -d postgres
	@echo "Waiting for postgres to be healthy..."
	@until docker compose exec postgres pg_isready -U chessmgr -q 2>/dev/null; do sleep 1; done
	@echo "Postgres is ready. Restart the backend to apply migrations."

# ── Code Generation ──────────────────────────────────────────────────────────
generate-backend: ## Generate backend server stubs from OpenAPI spec (requires oapi-codegen)
	cd backend && oapi-codegen --config oapi-codegen.yaml ../api/openapi.yaml

generate-frontend: ## Generate frontend API client from OpenAPI spec (requires openapi-typescript-codegen)
	cd frontend && npx openapi-typescript-codegen \
		--input ../api/openapi.yaml \
		--output src/api/generated \
		--client axios

generate-api: generate-backend generate-frontend ## Generate both backend and frontend API code
