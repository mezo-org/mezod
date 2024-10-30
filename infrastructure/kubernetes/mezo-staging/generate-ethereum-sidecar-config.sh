#!/bin/bash

read -p "ethereum node address: " ETHEREUM_NODE_ADDRESS

if [ -z "$ETHEREUM_NODE_ADDRESS" ]; then
  echo "ethereum node address cannot be empty"
  exit 1
fi

kubectl create secret generic \
  ethereum-sidecar-config \
  --from-literal=ethereum-node-address=$ETHEREUM_NODE_ADDRESS