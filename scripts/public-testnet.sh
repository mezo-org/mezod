#!/bin/bash

HOMEDIR=./.public-testnet

if [ -d "$HOMEDIR" ]; then
  echo "directory $HOMEDIR already exist; remove it to run this script"
  exit 1
fi

CHAIN_ID=mezo_31611-1
LIQUID_AMOUNT=1000000000000000000000abtc
NODE_DOMAIN=test.mezo.org
NODE_NAMES=("mezo-node-0" "mezo-node-1" "mezo-node-2" "mezo-node-3" "mezo-node-4" "mezo-faucet")

echo "using chain-id: $CHAIN_ID"
echo "using liquid amount: $LIQUID_AMOUNT"
echo "using node domain: $NODE_DOMAIN"
echo "using node names:" "${NODE_NAMES[@]}"

NODE_HOMEDIRS=()
NODE_ADDRESSES=()
NODE_MEMOS=()
KEYRING_PASSWORDS=()

for NODE_NAME in "${NODE_NAMES[@]}"; do
  NODE_HOMEDIR="$HOMEDIR/$NODE_NAME"
  NODE_KEY_NAME="$NODE_NAME-key"
  KEYRING_PASSWORD=$(openssl rand -hex 32)

  # Set some configuration options for the node to not repeat them in the commands.
  ./build/mezod --home=$NODE_HOMEDIR config set client chain-id $CHAIN_ID
  ./build/mezod --home=$NODE_HOMEDIR config set client keyring-backend file

  # Generate a new account key that will be used to authenticate blockchain transactions.
  # Capture the mnemonic used to generate that key.
  KEYS_ADD_OUT=$(yes $KEYRING_PASSWORD | ./build/mezod --home=$NODE_HOMEDIR keys add $NODE_KEY_NAME --output=json)
  MNEMONIC=$(echo $KEYS_ADD_OUT | jq -r '.mnemonic')

  KEYS_SHOW_OUT=$(yes $KEYRING_PASSWORD | ./build/mezod --home=$NODE_HOMEDIR keys show $NODE_KEY_NAME --output=json)
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
  yes "$MNEMONIC" | ./build/mezod --home=$NODE_HOMEDIR init $NODE_NAME --chain-id=$CHAIN_ID --recover &> /dev/null

  echo "[$NODE_NAME] init action done"

  # Generate validator data for the node using the genval command.
  # This step is skipped for mezo-faucet as this node is not a validator.
  if [ "$NODE_NAME" != "mezo-faucet" ]; then
    yes $KEYRING_PASSWORD | ./build/mezod --home=$NODE_HOMEDIR genval $NODE_KEY_NAME --ip="$NODE_NAME.$NODE_DOMAIN" &> /dev/null

    NODE_GENVAL=$(find $NODE_HOMEDIR/config/genval -mindepth 1 -print -quit)
    NODE_MEMOS+=($(jq -r '.memo' $NODE_GENVAL))

    echo "[$NODE_NAME] genval done"
  else
    echo "[$NODE_NAME] genval skipped"
  fi

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
  # Node's account balance must be added to the global genesis file explicitly.
  ./build/mezod --home=$GLOBAL_GENESIS_HOMEDIR add-genesis-account ${NODE_ADDRESSES[$i]} $LIQUID_AMOUNT &> /dev/null

  # Execute the rest for all nodes but the first.
  if [[ "$i" == '0' ]]; then
      continue
  fi

  NODE_HOMEDIR=${NODE_HOMEDIRS[$i]}

  # Move node's genval to the directory used to build the global genesis file.
  # We check if the node's genval directory exists as it might not if the node
  # is not a validator.
  NODE_GENVALDIR=$NODE_HOMEDIR/config/genval
  if [ -d "$NODE_GENVALDIR" ]; then
    mv $NODE_GENVALDIR/* $GLOBAL_GENESIS_HOMEDIR/config/genval
    rm -rf $NODE_GENVALDIR
  fi

  # Remove the local genesis file of the node as it won't be needed anymore.
  rm $NODE_HOMEDIR/config/genesis.json
done

# Aggregate all genval files into the global genesis file.
./build/mezod --home=$GLOBAL_GENESIS_HOMEDIR collect-genvals &> /dev/null
rm -rf $GLOBAL_GENESIS_HOMEDIR/config/genval

GENESIS=$GLOBAL_GENESIS_HOMEDIR/config/genesis.json
TMP_GENESIS=$GLOBAL_GENESIS_HOMEDIR/config/tmp_genesis.json

# Modify necessary parameters in the global genesis file
#
# [Modification 1]: Set abtc as the token denomination for relevant Cosmos SDK modules.
jq '.app_state["crisis"]["constant_fee"]["denom"]="abtc"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
# [Modification 2]: Set non-zero gas limit in genesis
jq '.consensus["params"]["block"]["max_gas"]="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
# [Modification 3]: Set the first node's address as the initial PoA owner.
POA_OWNER=${NODE_ADDRESSES[0]}
jq '.app_state["poa"]["owner"]="'"$POA_OWNER"'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Validate the global genesis file and move it to the root directory.
./build/mezod --home=$GLOBAL_GENESIS_HOMEDIR validate-genesis &> /dev/null
mv $GENESIS $HOMEDIR/genesis.json
GENESIS=$HOMEDIR/genesis.json # Reassign the GENESIS variable to the new location.

echo "global genesis file built and validated"

printf "%s\n" "${NODE_MEMOS[@]}" > $HOMEDIR/seeds.txt

for NODE_NAME in "${NODE_NAMES[@]}"; do
  NODE_HOMEDIR="$HOMEDIR/$NODE_NAME"
  NODE_CONFIGDIR="$NODE_HOMEDIR/config"
  NODE_APP_TOML="$NODE_CONFIGDIR/app.toml"
  NODE_CLIENT_TOML="$NODE_CONFIGDIR/client.toml"
  NODE_CONFIG_TOML="$NODE_CONFIGDIR/config.toml"

  # Cleanup the moniker from config. It will be set at startup using a flag.
  sed -i.bak 's/moniker = '\"$NODE_NAME\"'/moniker = ""/g' "$NODE_CONFIG_TOML"

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

  # Enable the necessary JSON-RPC namespaces to empower block explorers.
  sed -i.bak 's/api = "eth,net,web3"/api = "eth,net,web3,debug,miner,txpool,personal"/g' "$NODE_APP_TOML"

  # Set servers to listen on all interfaces, not just localhost.
  sed -i.bak 's/address = "127.0.0.1:8545"/address = "0.0.0.0:8545"/g' "$NODE_APP_TOML"
  sed -i.bak 's/ws-address = "127.0.0.1:8546"/ws-address = "0.0.0.0:8546"/g' "$NODE_APP_TOML"
  sed -i.bak 's/metrics-address = "127.0.0.1:6065"/metrics-address = "0.0.0.0:6065"/g' "$NODE_APP_TOML"
  sed -i.bak 's/node = "tcp:\/\/localhost:26657"/node = "tcp:\/\/0.0.0.0:26657"/g' "$NODE_CLIENT_TOML"
  sed -i.bak 's/proxy_app = "tcp:\/\/127.0.0.1:26658"/proxy_app = "tcp:\/\/0.0.0.0:26658"/g' "$NODE_CONFIG_TOML"
  sed -i.bak 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/g' "$NODE_CONFIG_TOML"
  sed -i.bak 's/pprof_laddr = "localhost:6060"/pprof_laddr = "0.0.0.0:6060"/g' "$NODE_CONFIG_TOML"

  # Remove all backup files created by sed.
  rm $NODE_CONFIGDIR/*.bak

  echo "[$NODE_NAME] configuration files prepared"
done