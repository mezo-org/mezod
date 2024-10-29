# Docker

## How to configure and run validator?

The following instruction will guide you through the process of configuring
and running a validator node. Before continuing, decide which network you want
to join. There are two options: `testnet` and `mainnet`. The following
instruction will use `testnet` as an example.

> [!NOTE]
> Run `make` (without arguments) to see the list of available commands.

#### Configuration and runtime flow

```mermaid
sequenceDiagram
  participant testnet.env
  participant Makefile
  box Docker Compose
    participant compose.yaml
    participant init as Service 'init'
    participant cli as Service 'cli'
    participant mezod as Service 'mezod' with sidecars
  end

  Note over testnet.env: (USER) Adjust the configuration
  Makefile -->> testnet.env: load

  critical One-time setup
    Makefile ->> compose.yaml: make init
    compose.yaml -->> init: run
    init -->> init: Load predefined variables (vars.sh)<br/>Generate mnemonic (optional)<br/>Prepare keyring<br/>Initialize the configuration
    Makefile ->> compose.yaml: make cli
    compose.yaml -->> cli: run
    Note over cli: (USER) Generate validator data
    Note over cli: (USER) Submit joining request (external)
  end

  loop Runtime
    Makefile ->> compose.yaml: make start
    compose.yaml -->> mezod: run
    Makefile ->> compose.yaml: make stop
    compose.yaml -->> mezod: stop
    Note over compose.yaml: (USER) Update the configuration
  end
```

### 1. Prepare configuration file

1. Copy the `testnet.env.example` to `testnet.env`:

```shell
cp testnet.env.example testnet.env
```
2. Edit the `testnet.env` file:
* `NETWORK` - the network you want to join (`testnet` or `mainnet`)
* `DOCKER_IMAGE` - the latest version of mezod image
* `LOCAL_BIND_PATH` - the path to the local directory where the data will be stored
* `KEYRING_PASSWORD` - the password for the keyring. It is used to encrypt the key.
Generate a new password using the following command:
```shell
$ openssl rand -hex 32
```
* `MEZOD_MONIKER` - the name of the validator
* `MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS` - the address of the Ethereum node


### 2. Initialize the configuration

```shell
make init
```

### 3. Generate validator data

```shell
make cli
$ echo "${KEYRING_PASSWORD}" | mezod genesis genval "${KEYRING_NAME}" --keyring-backend="file" --chain-id="${MEZOD_CHAIN_ID}" --home="${MEZOD_HOME}" --ip="${PUBLIC_IP}"
```

### 4. Submit joining request

TBD

### 5. Run the validator

```shell
make start
```
