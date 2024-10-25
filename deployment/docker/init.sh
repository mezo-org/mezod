#!/busybox/sh

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
    printf "Do you want to generate a new mnemonic? [y/N]"; read -r response
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
      --keyring-backend="${MEZOD_KEYRING_BACKEND}" \
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
    --keyring-backend="${MEZOD_KEYRING_BACKEND}" \
    --overwrite
  echo "Configuration initialized!"
}

validate_genesis() {
  echo "Validate genesis..."
  mezod genesis validate --home="${MEZOD_HOME}"
  echo "Genesis validated!"
}

#
# Main
#
prepare_keyring
init_configuration
validate_genesis
