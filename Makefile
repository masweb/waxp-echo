.PHONY: help dev build build-admin build-render run run-admin run-render migrate-up migrate-down sqlc generate tidy

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run admin with hot reload (requires air)
	air

build: build-admin build-render ## Build both binaries

build-admin: ## Build admin binary
	go build -o bin/admin ./cmd/admin

build-render: ## Build render binary
	go build -o bin/render ./cmd/render

run-admin: ## Run admin server
	go run ./cmd/admin

run-render: ## Run render server
	go run ./cmd/render

run: ## Run both servers
	@echo "Starting admin and render servers..."
	@$(MAKE) run-admin & $(MAKE) run-render & wait

tidy: ## Tidy modules
	go mod tidy

sqlc: ## Generate sqlc code
	sqlc generate

migrate-up: ## Run migrations up
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Run migrations down
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=create_posts_table)
	migrate create -ext sql -dir db/migrations -seq $(name)

generate: sqlc ## Generate all code
