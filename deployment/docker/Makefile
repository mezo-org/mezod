NETWORK ?= testnet
DOCKER_COMPOSE_CMD := docker compose --env-file $(NETWORK).env
DOCKER_COMPOSE_BACKUP_CMD := docker compose --env-file $(NETWORK).env --file compose.backup.yaml

include $(NETWORK).env


################################################################################
# Development
################################################################################

.PHONY: clean
clean: ## Clean up
	@rm -rf $(LOCAL_BIND_PATH)

.PHONY: build
build: ## Build Docker image
	docker build --platform linux/amd64 --tag $(DOCKER_IMAGE) ../../

.PHONY: cmd
cmd: ## Run command in management container (e.g., make cmd args="ls -la")
	$(DOCKER_COMPOSE_CMD) run --rm --interactive --entrypoint="/bin/sh -c" cli "$(args)"

################################################################################
# Initialization
################################################################################

.PHONY: setup
setup: # (no help) Setup local environment
	@mkdir -p $(LOCAL_BIND_PATH)
	@chmod 777 $(LOCAL_BIND_PATH)

.PHONY: init
init: setup ## Run init container
	@$(DOCKER_COMPOSE_CMD) run --rm --interactive init

.PHONY: cli
cli: ## Run shell in management container
	$(DOCKER_COMPOSE_CMD) run --rm --interactive cli

################################################################################
# Runtime
################################################################################

.PHONY: start
start: init ## Start node
	@$(DOCKER_COMPOSE_CMD) up --detach --force-recreate

.PHONY: stop
stop: ## Stop node
	$(DOCKER_COMPOSE_CMD) down

.PHONY: logs
logs: ## Show logs (e.g., make logs args="mezod")
	$(DOCKER_COMPOSE_CMD) logs --tail 50 --follow $(args)

################################################################################
# Help menu
################################################################################

.PHONY: help
help: ## Display help menu
	@echo "Usage: make [target]"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\t\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
