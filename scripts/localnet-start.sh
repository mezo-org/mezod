#!/bin/bash

HOMEDIR=./.localnet

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
NODE_HOMEDIR="$HOMEDIR/$NODE_NAME/evmosd"

echo "starting node $NODE_NAME with home directory $NODE_HOMEDIR"

./build/evmosd start --home "$NODE_HOMEDIR" \
  --json-rpc.api="eth,web3,net,debug,miner,txpool,personal" \
  --json-rpc.enable

echo "node $NODE_NAME started."
