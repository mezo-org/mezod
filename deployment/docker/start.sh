#!/busybox/sh

#
# This script starts the Mezod node.
#

set -o errexit # Exit on error
set -o nounset # Exit on use of an undefined variable

# Start the Mezod node
(echo ${KEYRING_PASSWORD}; echo ${KEYRING_PASSWORD}) \
  | mezod start \
    --log_format=${MEZOD_LOG_FORMAT} \
    --chain-id=${MEZOD_CHAIN_ID} \
    --home=${MEZOD_HOME} \
    --keyring-backend=${MEZOD_KEYRING_BACKEND} \
    --moniker=${MEZOD_MONIKER} \
    --ethereum-sidecar.client.server-address=${MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS} \
    --api.enable \
    --grpc.enable \
    --grpc-web.enable \
    --json-rpc.enable \
    --json-rpc.api eth,txpool,personal,net,debug,web3 \
    --metrics \
    --json-rpc.metrics-address="0.0.0.0:6065"
