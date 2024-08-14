# Precompile Hardhat Tasks

*Node.js v20 or greater is recommended*

Before getting started make sure you have changed to the hardhat directory and have installed the required
dependencies:

```
cd precompile/hardhat
npm install
```

## Networks

Hardhat is configured with two supported networks:

* `localhost` for connecting to local dev net (localnet-docker, localnet-bin).
* `mezo_testnet` for connecting to the public testnet.

All tasks and scripts support hardhat's global options which includes the `--network` flag. `localhost` is
used by default if no network is set. If running a task or script against `mezo_testnet` ensure you include
`--network mezo_testnet` as a hardhat argument.

## Accounts

Hardhat [vars](https://hardhat.org/hardhat-runner/docs/guides/configuration-variables) are used for configuring
accounts/keys.

```
WARNING

Configuration variables are stored in plain text on your disk. Avoid using this feature for data you wouldn’t
normally save in an unencrypted file. Run npx hardhat vars path to find the storage's file location.
```

You can determine what `vars` are available by running:

```
npx hardhat vars setup
```

The vars can be set to either a) a single private key, or b) a comma separated list (no whitespace) of private keys.

```
npx hardhat vars set MEZO_ACCOUNTS
```

And can be removed with:

```
npx hardhat vars delete MEZO_ACCOUNTS
```

## Scripts

Some helper scripts are included for convenience. It is recommended to run these scripts using `hardhat run` as they
may not execute as expected if running directly with node.

### accounts

Lists available accounts.

```
npx hardhat --network localhost run scripts/accounts.ts
```

### localhost-keys

Reads seed phrases from `build` dir used by localhost based localnets and prints a comma separated list of private
keys.

```
npx hardhat run scripts/localhost-keys.ts
```

This output can be used to easily set `MEZO_ACCOUNTS` for `localhost` use.

```
npx hardhat run scripts/localhost-keys.ts | npx hardhat vars set MEZO_ACCOUNTS
```

## Tasks

We name Hardhat [Tasks](https://hardhat.org/hardhat-runner/docs/advanced/create-task) with a precompile prefix. This
provides clarity on which precompile the task runs against, and keeps the output of `npx hardhat help` clean as
precompile tasks get grouped together. You can view the available tasks with:

```
npx hardhat help
```

*Note: The default Hardhat tasks are still visible (e.g `compile`), many of these will do nothing as we are only using
Hardhat for tasks/testing.*

Help information for a specific task can be obtained using

```
npx hardhat help <TASK>
```

e.g:

```
npx hardhat help validatorPool:validator
```

```
Usage: hardhat [GLOBAL OPTIONS] validatorPool:validator --operator <STRING>

OPTIONS:

  --operator	The validator's operator address

validatorPool:validator: Returns a validator's consensus public key & description
```

Here we can see the validatorPool:validator task has an operator argument - correct usage would be:

```
npx hardhat --network mezo_testnet validatorPool:validator --operator 0xc2f7Ae302a68CF215bb3dA243dadAB3290308015
```

### Running Tasks

Tasks get run as if they are built in hardhat commands. Read tasks are executed using a basic ethers provider
(no account), e.g:

```
npx hardhat --network mezo_testnet validatorPool:validators
```

Write tasks all require at minimum a signer argument, e.g:

```
npx hardhat --network mezo_testnet validatorPool:leave --signer 0xc2f7Ae302a68CF215bb3dA243dadAB3290308015
```

## Localnet

If running against a local dev net a minor adjustment to the genesis files is currently required. Due to this you must
clean and re-initialize:

```
make localnet-bin-clean
make localnet-bin-init
```

A small edit to the genesis files of each node then needs to be made before starting the network. In each nodes
`genesis.json` (`.localnet/node*/mezod/config/genesis.json`) update `consensus_params.blocks.max_gas` to `10000000`
(10 million). Then start the nodes as you normally would:

```
make localnet-bin-start
```

You can populate hardhat with the localnet accounts/private keys with the following

```
npx hardhat run scripts/localhost-keys.ts | npx hardhat vars set MEZO_ACCOUNTS
```
