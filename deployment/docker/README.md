# Docker

## How to configure and run validator?

The following instruction will guide you through the process of configuring
and running a validator node. Before continuing, decide which network you want
to join. There are two options: `testnet` and `mainnet`. For both networks,
we will use `NETWORK.env` file. Replace `NETWORK` with the desired network.

> [!NOTE]
> Run `make` to see the list of available commands.


### 1. Adjust configuration file

Adjust the configuration file `NETWORK.env`:
- `NETWORK` - network name (`testnet` or `mainnet`)
- `KEYRING_KEY_NAME` - keyring key name
- `KEYRING_MNEMONIC` - keyring mnemonic
- `KEYRING_PASSWORD` - keyring password
- `MEZOD_CHAIN_ID` - `mezo_31611-1` (for `testnet`) or `mainnet`
- `MEZOD_MONIKER` - moniker name
- `MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS` - Ethereum websocket endpoint

### 2. Initialize the configuration

```shell
make init
```

### 3. Generate validator data

```shell
make cli
$ mezod genesis genval "${MEZOD_MONIKER}" --keyring-backend="${KEYRING}" --chain-id="${CHAINID}" --home="${HOMEDIR}"
```

### 4. Submit joining request

TBD

### 5. Run the validator

```shell
make start
```
