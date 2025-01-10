#!/bin/bash

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

#start the ethereum sidecar
./build/mezod ethereum-sidecar \
  --ethereum-sidecar.server.ethereum-node-address=$ETH_SIDECAR_RPC_PROVIDER \
  --ethereum-sidecar.server.address=0.0.0.0:7500 &
ETH_SIDECAR_PID=$!

# start the connect sidecar
connect --market-map-endpoint=localhost:9090 &
CONNECT_SIDECAR_PID=$!


# catch ctrl-c
trap cleanup SIGINT

cleanup() {
  kill $ETH_SIDECAR_PID
  kill $CONNECT_SIDECAR_PID

  exit
}

wait
