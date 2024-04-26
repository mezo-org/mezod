#!/bin/bash

# TODO: Add faucet account to genesis.

HOMEDIR=./.public-testnet

if [ -d "$HOMEDIR" ]; then
  echo "directory $HOMEDIR already exist; remove it to run this script"
  exit 1
fi

CHAIN_ID=mezo_31611-1
STAKE_AMOUNT=100000000000000000000abtc
NODE_DOMAIN=test.mezo.org
NODE_NAMES=("mezo-node-0" "mezo-node-1" "mezo-node-2" "mezo-node-3" "mezo-node-4")

echo "using chain-id: $CHAIN_ID"
echo "using stake amount: $STAKE_AMOUNT"
echo "using node domain: $NODE_DOMAIN"
echo "using node names: ${NODE_NAMES[@]}"

NODE_HOMEDIRS=()
NODE_ADDRESSES=()
KEYRING_PASSWORDS=()

for NODE_NAME in "${NODE_NAMES[@]}"; do
  NODE_HOMEDIR="$HOMEDIR/$NODE_NAME"
  NODE_KEY_NAME="$NODE_NAME-key"
  KEYRING_PASSWORD=$(openssl rand -hex 32)

  # Set some configuration options for the node to not repeat them in the commands.
  ./build/evmosd --home=$NODE_HOMEDIR config chain-id $CHAIN_ID
  ./build/evmosd --home=$NODE_HOMEDIR config keyring-backend file

  # Generate a new account key that will be used to authenticate blockchain transactions.
  # Capture the mnemonic used to generate that key.
  KEYS_ADD_OUT=$(yes $KEYRING_PASSWORD | ./build/evmosd --home=$NODE_HOMEDIR keys add $NODE_KEY_NAME --output=json)
  MNEMONIC=$(echo $KEYS_ADD_OUT | jq -r '.mnemonic')

  KEYS_SHOW_OUT=$(yes $KEYRING_PASSWORD | ./build/evmosd --home=$NODE_HOMEDIR keys show $NODE_KEY_NAME --output=json)
  NODE_ADDRESS=$(echo $KEYS_SHOW_OUT | jq -r '.address')

  echo "[$NODE_NAME] account key generated"

  # By default, the init command generates:
  # - A new consensus key to participate in the CometBFT protocol
  # - A new network key for peer-to-peer authentication
  # Here we use the recover mode and generate the consensus key using the same
  # mnemonic as used for the account key. This makes management easier.
  # Worth noting that recover mode does not refer to the network key which is always
  # generated anew. However, the network key is not critical and can be replaced
  # without any consequences.
  yes "$MNEMONIC" | ./build/evmosd --home=$NODE_HOMEDIR init $NODE_NAME --chain-id=$CHAIN_ID --recover &> /dev/null

  echo "[$NODE_NAME] init action done"

  # Adding the account to the local node's genesis file is not strictly necessary
  # as this will be done for the global genesis file. However, it's needed to execute
  # the gentx command that checks the account balance in the local genesis file.
  yes $KEYRING_PASSWORD | ./build/evmosd --home=$NODE_HOMEDIR add-genesis-account $NODE_KEY_NAME $STAKE_AMOUNT &> /dev/null
  yes $KEYRING_PASSWORD | ./build/evmosd --home=$NODE_HOMEDIR gentx $NODE_KEY_NAME $STAKE_AMOUNT --ip="$NODE_NAME.$NODE_DOMAIN" &> /dev/null

  echo "[$NODE_NAME] gentx done"

  echo $KEYRING_PASSWORD > $NODE_HOMEDIR/keyring_password.txt
  echo $MNEMONIC > $NODE_HOMEDIR/mnemonic.txt

  echo "[$NODE_NAME] keyring password and mnemonic saved to files"

  NODE_HOMEDIRS+=("$NODE_HOMEDIR")
  NODE_ADDRESSES+=("$NODE_ADDRESS")
  KEYRING_PASSWORDS+=("$KEYRING_PASSWORD")
done

# Use the first node's homedir (and its local genesis file) to build the global genesis file.
GLOBAL_GENESIS_HOMEDIR=${NODE_HOMEDIRS[0]}

for i in "${!NODE_NAMES[@]}"; do
  # Execute for all nodes but the first.
  if [[ "$i" == '0' ]]; then
      continue
  fi

  NODE_HOMEDIR=${NODE_HOMEDIRS[$i]}

  # Move node's gentx to the directory used to build the global genesis file.
  # Remove the local genesis file of the node as it won't be needed anymore.
  mv $NODE_HOMEDIR/config/gentx/* $GLOBAL_GENESIS_HOMEDIR/config/gentx
  rm -rf $NODE_HOMEDIR/config/gentx
  rm $NODE_HOMEDIR/config/genesis.json

  # Node's account balance must be added to the global genesis file explicitly.
  ./build/evmosd --home=$GLOBAL_GENESIS_HOMEDIR add-genesis-account ${NODE_ADDRESSES[$i]} $STAKE_AMOUNT &> /dev/null
done

# Aggregate all gentx files into the global genesis file.
./build/evmosd --home=$GLOBAL_GENESIS_HOMEDIR collect-gentxs &> /dev/null
rm -rf $GLOBAL_GENESIS_HOMEDIR/config/gentx

# Validate the global genesis file and move it to the root directory.
./build/evmosd --home=$GLOBAL_GENESIS_HOMEDIR validate-genesis &> /dev/null
mv $GLOBAL_GENESIS_HOMEDIR/config/genesis.json $HOMEDIR/genesis.json

echo "global genesis file built and validated"

SEEDS=$(jq -r '.app_state.genutil.gen_txs | .[] | .body.memo' $HOMEDIR/genesis.json)
printf "%s\n" "${SEEDS[@]}" > $HOMEDIR/seeds.txt


for NODE_NAME in "${NODE_NAMES[@]}"; do
  NODE_HOMEDIR="$HOMEDIR/$NODE_NAME"
  NODE_CONFIGDIR="$NODE_HOMEDIR/config"
  NODE_APP_TOML="$NODE_CONFIGDIR/app.toml"
  NODE_CONFIG_TOML="$NODE_CONFIGDIR/config.toml"

  # All initial validators should maintain connections to each other.
  # This is why the seeds.txt is used to populate the persistent_peers field
  # and the seeds field is emptied due to being redundant.
  sed -i.bak -e "s/^seeds =.*/seeds = \"\"/" $NODE_CONFIG_TOML
  sed -i.bak -e "s/^persistent_peers =.*/persistent_peers = \"$(paste -s -d, $HOMEDIR/seeds.txt)\"/" $NODE_CONFIG_TOML

  # Set the default minimum gas prices to 1 satoshi.
  sed -i.bak 's/minimum-gas-prices = "0abtc"/minimum-gas-prices = "10000000000abtc"/g' "$NODE_APP_TOML"

  # Set the pruning mode to nothing to make this an archiving node
  sed -i.bak 's/pruning = "default"/pruning = "nothing"/g' "$NODE_APP_TOML"

  # Enable Prometheus metrics.
  sed -i.bak 's/prometheus = false/prometheus = true/' "$NODE_CONFIG_TOML"
  sed -i.bak 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$NODE_APP_TOML"
  sed -i.bak 's/enabled = false/enabled = true/g' "$NODE_APP_TOML"

  # Remove all backup files created by sed.
  rm $NODE_CONFIGDIR/*.bak

  echo "[$NODE_NAME] configuration files prepared"
done