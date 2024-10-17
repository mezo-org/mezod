DOCKER_COMPOSE_CMD := docker compose --env-file monitoring.env

.PHONY: clean start stop logs raw help

include monitoring.env

clean: ## Clean up
	rm -rf ${STORAGE}/*

start: ## Start node
	$(DOCKER_COMPOSE_CMD) up --detach --force-recreate

stop: ## Stop node
	$(DOCKER_COMPOSE_CMD) down

logs: ## Show logs
	$(DOCKER_COMPOSE_CMD) logs --tail 50 --follow $(args)

raw: ## Wrapper for Docker Compose
	$(DOCKER_COMPOSE_CMD) $(args)

watch: ## Watch containers
	watch -n 1 $(DOCKER_COMPOSE_CMD) ps

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\t\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
