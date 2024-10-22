#!/busybox/sh

#
# This script initializes the Mezod node configuration and keyring.
#

set -o errexit # Exit on error
set -o nounset # Exit on use of an undefined variable

prepare_keyring() {
  echo "Prepare keyring..."
  (echo ${KEYRING_MNEMONIC}; echo ${KEYRING_PASSWORD}; echo ${KEYRING_PASSWORD}) \
    | mezod keys add \
      ${KEYRING_KEY_NAME} \
      --home=${MEZOD_HOME} \
      --keyring-backend=${MEZOD_KEYRING_BACKEND} \
      --recover
  echo "Keyring prepared!"
}

init_configuration() {
  echo "Initialize configuration..."
  mezod \
    init \
    ${MEZOD_MONIKER} \
    --chain-id=${MEZOD_CHAIN_ID} \
    --home=${MEZOD_HOME} \
    --keyring-backend=${MEZOD_KEYRING_BACKEND} \
    --overwrite
  echo "Configuration initialized!"
}

validate_genesis() {
  echo "Validate genesis..."
  mezod genesis validate --home=${MEZOD_HOME}
  echo "Genesis validated!"
}

#
# Main
#
prepare_keyring
init_configuration
validate_genesis
