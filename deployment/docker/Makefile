NETWORK ?= testnet
DOCKER_COMPOSE_CMD := docker compose --env-file $(NETWORK).env
cmd ?= ps

.PHONY: clean build init help start stop logs raw

clean: ## Clean up
	rm -rf tmp/storage

build: ## Build Docker image
	docker build --platform linux/amd64 -t mezod ../../

setup: ## Setup configuration
	mkdir -p tmp/storage
	$(DOCKER_COMPOSE_CMD) run --rm init
	$(DOCKER_COMPOSE_CMD) run --rm genesis

start: ## Start node
	$(DOCKER_COMPOSE_CMD) up --detach

stop: ## Stop node
	$(DOCKER_COMPOSE_CMD) down

logs: ## Show logs
	$(DOCKER_COMPOSE_CMD) logs --tail 50 --follow $(args)

raw: ## Wrapper for Docker Compose
	$(DOCKER_COMPOSE_CMD) $(args)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\t\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
