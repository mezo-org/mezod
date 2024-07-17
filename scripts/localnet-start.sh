#!/bin/bash

HOMEDIR=./.localnet
CHAIN_ID=mezo_31611-1
LOGLEVEL="info"
TRACE=""

if [ ! -d "$HOMEDIR" ]; then
  echo "localnet directory $HOMEDIR does not exist; run scripts/localnet.sh first."
  exit 1
fi

NODE_NAMES=($(ls -d $HOMEDIR/mezo-node-* | xargs -n 1 basename))

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
NODE_HOMEDIR="$HOMEDIR/$NODE_NAME"

echo "starting node $NODE_NAME with home directory $NODE_HOMEDIR"

./build/evmosd start --home "$NODE_HOMEDIR" --log_level "$LOGLEVEL" $TRACE

echo "node $NODE_NAME started."
