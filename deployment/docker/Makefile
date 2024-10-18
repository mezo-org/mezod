NETWORK ?= testnet
DOCKER_COMPOSE_CMD := docker compose --env-file $(NETWORK).env

include $(NETWORK).env

.PHONY: clean build init help start stop logs raw setup

clean: ## Clean up
	rm -rf tmp/mezod
	# rm -rf tmp/keyring

build: ## Build Docker image
	docker build --platform linux/amd64 --tag $(DOCKER_IMAGE) ../../

setup: ## Setup local environment
	mkdir -p tmp/mezod
	# mkdir -p tmp/keyring

start-one: setup ## Run single service
	@test $(args) || (echo "Usage: make start-one args=<service>" && exit 1)
	$(DOCKER_COMPOSE_CMD) run --rm $(args)

start: setup ## Start node
	$(DOCKER_COMPOSE_CMD) up --detach --force-recreate

stop: ## Stop node
	$(DOCKER_COMPOSE_CMD) down

logs: ## Show logs
	$(DOCKER_COMPOSE_CMD) logs --tail 50 --follow $(args)

raw: ## Wrapper for Docker Compose
	$(DOCKER_COMPOSE_CMD) $(args)

cli: ## Run management container
	$(DOCKER_COMPOSE_CMD) run --rm --interactive cli

help: ## Display help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\t\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
