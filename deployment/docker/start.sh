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
    --api.address="tcp://0.0.0.0:1317" \
    --grpc.enable \
    --grpc.address="0.0.0.0:9090" \
    --grpc-web.enable \
    --json-rpc.enable \
    --json-rpc.address="0.0.0.0:8545" \
    --json-rpc.api eth,txpool,personal,net,debug,web3 \
    --json-rpc.ws-address="0.0.0.0:8546" \
    --metrics \
    --json-rpc.metrics-address="10.55.0.6:6065" \
    --p2p.laddr="tcp://0.0.0.0:26656" \
    --node="tcp://0.0.0.0:26657"
