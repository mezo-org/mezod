#!/bin/bash

HOMEDIR=./.localnet

# load custom configuration via the .env file
# for now a single option is available to set the ethereum
# RPC provider URL, e.g:
# ETH_SIDECAR_RPC_PROVIDER=wss://my.eth.rpc/provider
if [ -f .env ]; then
  set -a; source .env; set +a

  if [ -z "${ETH_SIDECAR_RPC_PROVIDER}" ]; then
    echo "warning: no ETH_SIDECAR_RPC_PROVIDER variable set in .env file, ethereum sidecar may not start successfully."
  fi
else
  echo "warning: no .env file specified, some binary might not start successfully."
fi

if [ -z "${LOCALNET_CHAIN_ID}" ]; then
  echo "Please set the LOCALNET_CHAIN_ID environment variable."
  exit 1
fi

if [ ! -d "$HOMEDIR" ]; then
  echo "localnet directory $HOMEDIR does not exist; run make localnet-bin-init first."
  exit 1
fi

NODE_NAMES=()
while IFS= read -r -d '' dir; do
  NODE_NAMES+=("$(basename "$dir")")
done < <(find "$HOMEDIR" -maxdepth 1 -type d -name 'node*' -print0 | sort -z)

if [ ${#NODE_NAMES[@]} -eq 0 ]; then
  echo "No nodes found."
  exit 1
fi

echo "available nodes:"
for i in "${!NODE_NAMES[@]}"; do
  echo "[$i] ${NODE_NAMES[$i]}"
done

echo "select node to start by entering the corresponding number:"
read -r NODE_INDEX

if ! [[ "$NODE_INDEX" =~ ^[0-9]+$ ]] || [ "$NODE_INDEX" -ge "${#NODE_NAMES[@]}" ]; then
  echo "invalid selection; please enter a number between 0 and $((${#NODE_NAMES[@]} - 1))."
  exit 1
fi

NODE_NAME=${NODE_NAMES[$NODE_INDEX]}
NODE_HOMEDIR="$HOMEDIR/$NODE_NAME/mezod"

echo "starting node $NODE_NAME with home directory $NODE_HOMEDIR"

# create a port per eth sidecar per node
ETH_SIDECAR_PORT=$((7500 + $NODE_INDEX))
ETH_SIDECAR_ADDRESS="0.0.0.0:$ETH_SIDECAR_PORT"

#start the ethereum sidecar
./build/mezod ethereum-sidecar \
  --ethereum-sidecar.server.ethereum-node-address=$ETH_SIDECAR_RPC_PROVIDER \
  --ethereum-sidecar.server.address=$ETH_SIDECAR_ADDRESS &
ETH_SIDECAR_PID=$!

# create a port per connect sidecar per node
CONNECT_SIDECAR_PORT=$((7600 + $NODE_INDEX))
CONNECT_SIDECAR_ADDRESS="0.0.0.0:$CONNECT_SIDECAR_PORT"

# the mezo node grpc address, required for the market map call from the connect sidecar
MEZOD_GRPC_PORT=$((9200 + $NODE_INDEX))
MEZOD_GRPC_ADDRESS="localhost:$MEZOD_GRPC_PORT"

# start the connect sidecar
connect --port=$CONNECT_SIDECAR_PORT --market-map-endpoint=$MEZOD_GRPC_ADDRESS &
CONNECT_SIDECAR_PID=$!

# start the mezod binary
./build/mezod start --home "$NODE_HOMEDIR" \
  --chain-id=$LOCALNET_CHAIN_ID \
  --json-rpc.api="eth,web3,net,debug,miner,txpool,personal" \
  --json-rpc.enable \
  --ethereum-sidecar.client.server-address=$ETH_SIDECAR_ADDRESS \
  --oracle.oracle_address=$CONNECT_SIDECAR_ADDRESS \
  --grpc.address=$MEZOD_GRPC_ADDRESS &

MEZOD_PID=$!

echo "node $NODE_NAME started."

# catch ctrl-c
trap cleanup SIGINT

cleanup() {
  kill $ETH_SIDECAR_PID
  kill $CONNECT_SIDECAR_PID
  kill $MEZOD_PID
  exit
}

wait
