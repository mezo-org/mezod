#!/usr/bin/env bash

set -o errexit  # Exit on error
set -o nounset  # Exit on using unset variable
set -o pipefail # Exit on pipe error


################################################################################
# Variables
################################################################################
NETWORK="${NETWORK:-testnet}"
DOCKER_COMPOSE_CMD="docker compose --env-file ${NETWORK}.env"

# Load NETWORK.env
if [ -f "${NETWORK}.env" ]; then
  # shellcheck source=/dev/null
  source "${NETWORK}.env"
else
  echo "Error: ${NETWORK}.env file not found!"
  exit 1
fi

################################################################################
# Helper functions
################################################################################
_build_cli_image() {
  echo "Building Validator Kit CLI image..."
  docker build --platform linux/amd64 --tag local/mezod-cli --file - . <<EOF 2>/dev/null
FROM ${DOCKER_IMAGE} AS mezod
FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    jq \
    curl \
    nano \
    vim \
    && rm -rf /var/lib/apt/lists/*

# Add nonroot group used in the mezod image
RUN groupadd --gid 65532 nonroot

# Add nonroot user used in the mezod image
RUN useradd --uid 65532 --gid 65532 --no-create-home nonroot

COPY --from=mezod /entrypoint.sh /entrypoint.sh
COPY --from=mezod /usr/bin/mezod /usr/bin/mezod

USER nonroot
EOF
}

_run_cli_cmd_oneshoot() {
  _build_cli_image
  $DOCKER_COMPOSE_CMD run --rm --no-TTY cli /bin/bash
}

################################################################################
# Development and Operations
################################################################################
clean() { # (no help) Remove all data
  _run_cli_cmd_oneshoot <<EOF
set -x
rm -rf "${MEZOD_HOME}/*"
EOF
}

reset() { # (no help) Reset state
  _run_cli_cmd_oneshoot <<EOF
set -x
rm -rf "${MEZOD_HOME}"/data/*.db
rm -rf "${MEZOD_HOME}"/config/addrbook.json
rm -rf "${MEZOD_HOME}"/config/write-file-atomic-*
EOF
}

validator-info() { ## Show validator information
  _run_cli_cmd_oneshoot <<EOF
/bin/bash /entrypoint.sh info
EOF
}

################################################################################
# Initialization
################################################################################
init-keyring() { ## Initialize keyring
  ${DOCKER_COMPOSE_CMD} run --no-deps --rm --interactive mezod keyring
}

init-config() { ## Initialize configuration
  ${DOCKER_COMPOSE_CMD} run --no-deps --rm --interactive mezod config
}

init-genval() { ## Generate validator key
  _run_cli_cmd_oneshoot <<'EOF'
echo "${KEYRING_PASSWORD}" | mezod genesis genval "${KEYRING_NAME}" --keyring-backend="file" --chain-id="${MEZOD_CHAIN_ID}" --home="${MEZOD_HOME}" --ip="${PUBLIC_IP}"
EOF
}

################################################################################
# Runtime
################################################################################
start() { ## Start the node
  ${DOCKER_COMPOSE_CMD} up --detach --force-recreate
}

stop() { ## Stop the node
  ${DOCKER_COMPOSE_CMD} down
}

logs() { ## Show logs
  ${DOCKER_COMPOSE_CMD} logs --follow --tail 100 "$@"
}

journal() { ## Show journalctl logs
  journalctl CONTAINER_NAME="$1" -f
}

################################################################################
# Help
################################################################################
help() { ## Show help
  echo "Usage: $0 <command>"
  grep -E '^[a-zA-Z0-9_\-]+ *\(\) *{ *##' "$0" \
    | sed 's/() *{ *## */ - /g' \
    | awk '{printf "\t%-30s %s\n", $1, substr($0, index($0, $3))}'
}

################################################################################
# Main
################################################################################
if [ $# -eq 0 ]; then
  # Show help if no argument is provided
  help
  exit 1
else
  # Run the command
  "$@"
fi
