#!/bin/bash

ROOT_DIR="$(realpath "$(dirname $0)/../../../")"

NODE_NAME_DEFAULT=mezo-node-0
read -p "node name [$NODE_NAME_DEFAULT]: " NODE_NAME
NODE_NAME=${NODE_NAME:-$NODE_NAME_DEFAULT}

NODE_HOMEDIR_DEFAULT=$ROOT_DIR/.public-testnet/$NODE_NAME
read -p "node config homedir [$NODE_HOMEDIR_DEFAULT]: " NODE_HOMEDIR
NODE_HOMEDIR=${NODE_HOMEDIR:-$NODE_HOMEDIR_DEFAULT}

if [ ! -d "$NODE_HOMEDIR" ]; then
  echo "directory $NODE_HOMEDIR does not exist; you may need to run the public testnet artifact generator first (scripts/public-testnet.sh)"
  exit 1
fi

# This script assumes that the node's configuration has been generated using the
# scripts/public-testnet.sh script and walk though the directory structure
# produced by that script to create a Kubernetes secret with the node's keystore.
kubectl create secret generic \
  $NODE_NAME-keystore \
  --from-file=$NODE_HOMEDIR/config/node_key.json \
  --from-file=$NODE_HOMEDIR/config/priv_validator_key.json \
  --from-file=$NODE_HOMEDIR/keyring-file \
  --from-file=$NODE_HOMEDIR/keyring_password.txt \
  --from-file=$NODE_HOMEDIR/mnemonic.txt
