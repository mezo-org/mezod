NETWORK ?= testnet
DOCKER_COMPOSE_CMD := docker compose --env-file $(NETWORK).env
DOCKER_COMPOSE_BACKUP_CMD := docker compose --env-file $(NETWORK).env --file compose.backup.yaml

include $(NETWORK).env


################################################################################
# Management
################################################################################

.PHONY: clean
clean: ## Clean up
	@rm -rf $(MOUNTED_DIR)

.PHONY: build
build: ## Build Docker image
	docker build --platform linux/amd64 --tag $(DOCKER_IMAGE) ../../


################################################################################
# Runtime
################################################################################

.PHONY: setup
setup: # (no help) Setup local environment
	@mkdir -p $(MOUNTED_DIR)
	@chmod 777 $(MOUNTED_DIR)

.PHONY: init
init: setup ## Run init container
	@$(DOCKER_COMPOSE_CMD) run --rm --interactive init

.PHONY: start
start: init ## Start node
	@$(DOCKER_COMPOSE_CMD) up --detach --force-recreate

.PHONY: stop
stop: ## Stop node
	$(DOCKER_COMPOSE_CMD) down

.PHONY: logs
logs: ## Show logs
	$(DOCKER_COMPOSE_CMD) logs --tail 50 --follow $(args)

.PHONY: cli
cli: ## Run shell in management container
	$(DOCKER_COMPOSE_CMD) run --rm --interactive cli

.PHONY: cmd
cmd: ## Run command in management container
	$(DOCKER_COMPOSE_CMD) run --rm --interactive --entrypoint="/busybox/sh -c" cli "$(args)"


################################################################################
# Backup
################################################################################

.PHONY: backup-create
backup-create: ## Create a backup
	@mkdir -p ./tmp/backup
	$(DOCKER_COMPOSE_BACKUP_CMD) run --rm --interactive backup-create

.PHONY: backup-list
backup-list: ## List backups
	$(DOCKER_COMPOSE_BACKUP_CMD) run --rm --interactive backup-list

.PHONY: backup-restore
backup-restore: ## Restore a backup
	@mkdir -p ./tmp/backup-restore
	$(DOCKER_COMPOSE_BACKUP_CMD) run --rm --interactive backup-restore


################################################################################
# Help menu
################################################################################

.PHONY: help
help: ## Display help menu
	@echo "Usage: make [target]"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\t\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
