.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Cluster management
cluster-create: ## Create k3d cluster
	@echo "Creating k3d cluster..."
	# Implementation here

cluster-delete: ## Delete k3d cluster
	@echo "Deleting k3d cluster..."
	# Implementation here

# Deployment
deploy-all: deploy-kafka deploy-clickhouse deploy-minio deploy-vector deploy-grafana deploy-prometheus deploy-api ## Deploy all components

deploy-kafka: ## Deploy Kafka (Strimzi + KRaft)
	@echo "Deploying Kafka..."
	# Implementation here

deploy-clickhouse: ## Deploy ClickHouse
	@echo "Deploying ClickHouse..."
	# Implementation here

deploy-minio: ## Deploy MinIO
	@echo "Deploying MinIO..."
	# Implementation here

deploy-vector: ## Deploy Vector
	@echo "Deploying Vector..."
	# Implementation here

deploy-grafana: ## Deploy Grafana
	@echo "Deploying Grafana..."
	# Implementation here

deploy-prometheus: ## Deploy Prometheus
	@echo "Deploying Prometheus..."
	# Implementation here

deploy-api: ## Deploy Query API
	@echo "Deploying API..."
	# Implementation here

# Docker Compose
compose-up: ## Start all services with Docker Compose
	docker-compose up -d

compose-down: ## Stop all services
	docker-compose down -v

compose-dev: ## Start dev services only
	docker-compose up -d kafka clickhouse vector

# Development
api-build: ## Build Go API
	go build -o bin/api ./cmd/api

api-test: ## Run API tests
	go test -v ./...

api-run: api-build ## Run API locally
	./bin/api

# Database
db-migrate: ## Run ClickHouse migrations
	@echo "Running DB migrations..."
	# Implementation here

db-shell: ## Open ClickHouse shell
	docker exec -it clickhouse clickhouse-client

# Testing
generate-logs: ## Generate test logs
	@echo "Generating test logs..."
	# Implementation here

load-test: ## Run k6 load tests
	k6 run scripts/load_test.js

chaos-test: ## Run chaos tests
	@echo "Running chaos tests..."
	# Implementation here

# Utilities
port-forward: ## Port-forward all services
	@echo "Port forwarding services..."
	# Implementation here

logs: ## View component logs
	@echo "Viewing logs..."
	# Implementation here

status: ## Show cluster status
	kubectl get pods -A

clean: ## Clean everything
	rm -rf bin/
	# Other cleanup tasks
