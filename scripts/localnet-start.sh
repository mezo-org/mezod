#!/bin/bash

HOMEDIR=./.localnet

if [ ! -d "$HOMEDIR" ]; then
  echo "localnet directory $HOMEDIR does not exist; run localnet-bin-init first."
  exit 1
fi

NODE_NAMES=()
while IFS= read -r -d '' dir; do
  NODE_NAMES+=("$(basename "$dir")")
done < <(find "$HOMEDIR" -maxdepth 1 -type d -name 'node*' -print0)

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

# Choose the RPC and WS ports based on the node index.
# The node at index 0 should use ports 8545 and 8546.
# The node at index 1 should use ports 8547 and 8548. And so on.
RPC_ADDRESS="0.0.0.0:$((8545 + NODE_INDEX * 2))"
WS_ADDRESS="0.0.0.0:$((8546 + NODE_INDEX * 2))"

echo "starting node $NODE_NAME with home directory $NODE_HOMEDIR and addresses:"
echo "--json-rpc.address=\"$RPC_ADDRESS\" --json-rpc.ws-address=\"$WS_ADDRESS\""

./build/evmosd start --home "$NODE_HOMEDIR" \
  --json-rpc.address="$RPC_ADDRESS" \
  --json-rpc.ws-address="$WS_ADDRESS" \
  --json-rpc.api="eth,web3,net,debug,miner,txpool,personal" \
  --json-rpc.enable

echo "node $NODE_NAME started."
