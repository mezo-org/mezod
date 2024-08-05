#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
# invoke like: bash scripts/ss.bash



## USAGE RUNDOWN
# Not for use on live nodes
# For use when testing.
# Assumes that ~/.mezod doesn't exist
# can be modified to suit your purposes if ~/.mezod does already exist


set -uxe

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

go install ./...

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...

# Initialize chain.
mezod init test --chain-id mezo_31611-1

# Get Genesis
wget https://archive.mezo.org/mainnet/genesis.json
mv genesis.json ~/.mezod/config/


# Get "trust_hash" and "trust_height".
INTERVAL=1000
LATEST_HEIGHT=$(curl -s https://mezo-rpc.polkachu.com/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "https://mezo-rpc.polkachu.com/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export MEZOD_STATESYNC_ENABLE=true
export MEZOD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export MEZOD_STATESYNC_RPC_SERVERS="https://rpc.mezo.interbloc.org:443,https://mezo-rpc.polkachu.com:443,https://tendermint.bd.mezo.org:26657,https://rpc.mezo.posthuman.digital:443,https://rpc.mezo.testnet.run:443,https://rpc.mezo.bh.rocks:443"
export MEZOD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export MEZOD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
export MEZOD_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/mezo/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')

# Start chain.
mezod start --x-crisis-skip-assert-invariants 
