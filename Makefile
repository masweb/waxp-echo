.PHONY: help dev build run migrate-up migrate-down sqlc generate tidy

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Run with hot reload (requires air)
	air

build: ## Build binary
	go build -o bin/server ./cmd/server

run: ## Run server
	go run ./cmd/server

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
