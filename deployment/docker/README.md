# Docker

## How to configure and run validator?

The following instruction will guide you through the process of configuring
and running a validator node. Before continuing, decide which network you want
to join. There are two options: `testnet` and `mainnet`. The following
instruction will use `testnet` as an example.

> [!NOTE]
> Run `make` (without arguments) to see the list of available commands.

### 1. Adjust configuration file

Adjust the configuration file `testnet.env`:

- `NETWORK` - network name (`testnet` or `mainnet`)
- `KEYRING_KEY_NAME` - keyring key name
- `KEYRING_MNEMONIC` - [keyring mnemonic](#how-to-generate-a-new-mnemonic)
- `KEYRING_PASSWORD` - keyring password
- `MEZOD_CHAIN_ID` - `mezo_31611-1` (for `testnet`) or `mainnet`
- `MEZOD_MONIKER` - moniker name
- `MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS` - Ethereum websocket endpoint

#### How to generate a new mnemonic?

```shell
NETWORK=testnet make cli
$ mezod keys mnemonic
```

### 2. Initialize the configuration

```shell
NETWORK=testnet make init
```

### 3. Generate validator data

```shell
NETWORK=testnet make cli
$ echo "${KEYRING_PASSWORD}" | mezod genesis genval "${MEZOD_MONIKER}" --keyring-backend="${MEZOD_KEYRING_BACKEND}" --chain-id="${MEZOD_CHAIN_ID}" --home="${MEZOD_HOME}"
```

### 4. Submit joining request

TBD

### 5. Run the validator

```shell
NETWORK=testnet make start
```
