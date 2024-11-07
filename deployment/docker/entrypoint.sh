#!/bin/sh

#
# This script initializes the Mezod node configuration and keyring.
#

set -o errexit # Exit on error

gen_mnemonic() {
  mnemonic_file="$1"
  # Check if the environment variable is set
  if [ -n "$KEYRING_MNEMONIC" ]; then
    echo "Using KEYRING_MNEMONIC from the environment"
    echo "$KEYRING_MNEMONIC" > "$mnemonic_file"
  else
    # Ask the user to generate a new mnemonic
    printf "Do you want to generate a new mnemonic? [y/N]: "; read -r response
    case "$response" in
      [yY])
        echo "Generating a new mnemonic..."
        m=$(mezod keys mnemonic)
        echo "$m" > "$mnemonic_file"
        printf "\n%s\n%s\n\n" "Generated mnemonic (make backup!):" "$m"
        echo "Press any key to continue..."; read -r _
        ;;
      *)
        # Ask the user to enter the mnemonic
        printf "Enter the mnemonic: "; read -r mnemonic
        echo "$mnemonic" > "$mnemonic_file"
        ;;
    esac
  fi
}

prepare_keyring() {
  test -f "${MEZOD_HOME}/keyring-file/keyhash" && {
    echo "Keyring already prepared!"
    return
  }

  mnemonic_file="/tmp/mnemonic.txt"
  gen_mnemonic "${mnemonic_file}"
  read -r keyring_mnemonic < "${mnemonic_file}"

  echo "Prepare keyring..."
  (echo "${keyring_mnemonic}"; echo "${KEYRING_PASSWORD}"; echo "${KEYRING_PASSWORD}") \
    | mezod keys add \
      "${KEYRING_NAME}" \
      --home="${MEZOD_HOME}" \
      --keyring-backend="file" \
      --recover
  echo "Keyring prepared!"
}

init_configuration() {
  echo "Initialize configuration..."
  mezod \
    init \
    "${MEZOD_MONIKER}" \
    --chain-id="${MEZOD_CHAIN_ID}" \
    --home="${MEZOD_HOME}" \
    --keyring-backend="file" \
    --overwrite
  echo "Configuration initialized!"
}

validate_genesis() {
  echo "Validate genesis..."
  mezod genesis validate --home="${MEZOD_HOME}"
  echo "Genesis validated!"
}

customize_configuration() {
  client_config_file="${MEZOD_HOME}/config/client.toml"
  app_config_file="${MEZOD_HOME}/config/app.toml"
  config_file="${MEZOD_HOME}/config/config.toml"

  echo "Backup original configuration..."
  test -f "${client_config_file}.bak" || cat "$client_config_file" > "${client_config_file}.bak"
  test -f "${app_config_file}.bak" || cat "$app_config_file" > "${app_config_file}.bak"
  test -f "${config_file}.bak" || cat "$config_file" > "${config_file}.bak"

  echo "Customize configuration..."

  #
  # FILE: client.toml
  #
  mezod toml set "$client_config_file" \
    -v chain-id="${MEZOD_CHAIN_ID}" \
    -v keyring-backend="file"

  #
  # FILE: config.toml
  #
  mezod toml set "$config_file" \
    -v moniker="${MEZOD_MONIKER}" \
    -v p2p.laddr="tcp://0.0.0.0:26656" \
    -v p2p.external_address="${PUBLIC_IP}:26656" \
    -v rpc.laddr="tcp://0.0.0.0:26657" \
    -v instrumentation.prometheus="true" \
    -v instrumentation.prometheus_listen_addr="0.0.0.0:26660"


  # Increase timeouts
  mezod toml set "$config_file" \
    -v consensus.timeout_propose="30s" \
    -v consensus.timeout_propose_delta="5s" \
    -v consensus.timeout_prevote="10s" \
    -v consensus.timeout_prevote_delta="5s" \
    -v consensus.timeout_precommit="5s" \
    -v consensus.timeout_precommit_delta="5s" \
    -v consensus.timeout_commit="150s" \
    -v rpc.timeout_broadcast_tx_commit="150s"

  #
  # FILE: app.toml
  #
  mezod toml set "$app_config_file" \
    -v ethereum-sidecar.client.server-address="ethereum-sidecar:7500" \
    -v api.enable="true" \
    -v api.address="tcp://0.0.0.0:1317" \
    -v grpc.enable="true" \
    -v grpc.address="0.0.0.0:9090" \
    -v grpc-web.enable="true" \
    -v json-rpc.enable="true" \
    -v json-rpc.address="0.0.0.0:8545" \
    -v json-rpc.api="eth,txpool,personal,net,debug,web3" \
    -v json-rpc.ws-address="0.0.0.0:8546" \
    -v json-rpc.metrics-address="0.0.0.0:6065"


  echo "Configuration customized!"
}

get_validator_info() {
  validator_addr_bech="$(echo "${KEYRING_PASSWORD}" | mezod --home="${MEZOD_HOME}" keys show "${KEYRING_NAME}" --address)"
  validator_addr="$(mezod --home="${MEZOD_HOME}" keys parse "${validator_addr_bech}" | grep bytes | awk '{print "0x"$2}')"
  echo "Validator address: ${validator_addr}"

  validator_id="$(cat "${MEZOD_HOME}"/config/genval/genval-*.json | jq -r '.memo' | awk -F'@' '{print $1}')"
  echo "Validator ID: ${validator_id}"

  validator_consensus_addr_bech="$(cat "${MEZOD_HOME}"/config/genval/genval-*.json | jq -r '.validator.cons_pub_key_bech32')"
  validator_consensus_addr="$(mezod --home="${MEZOD_HOME}" keys parse "${validator_consensus_addr_bech}" | grep bytes | awk '{printf "%s", $2}' | tail -c 64 | awk '{print "0x"$1}')"
  echo "Validator consensus address: ${validator_consensus_addr}"

  validator_network_addr="$(jq -r '.address' "${MEZOD_HOME}"/config/priv_validator_key.json | awk '{print "0x"$1}')"
  echo "Validator network address: ${validator_network_addr}"
}

#
# MAIN
#
if [ -z "$1" ]; then
  echo "No command provided!"
  exit 1
fi

case "$1" in
  keyring)
    prepare_keyring
    exit 0
    ;;
  info)
    get_validator_info
    exit 0
    ;;
  config)
    init_configuration
    validate_genesis
    customize_configuration
    exit 0
    ;;
  *)
    init_configuration
    validate_genesis
    customize_configuration
    # Run the mezod node
    exec "$@"
    ;;
esac
