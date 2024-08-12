# Precompile Hardhat Tasks

## Networks
Hardhat is configured with two supported networks:
* `localhost` for connecting to local dev net (localnet-docker, localnet-bin).
* `mezo_testnet` for connecting to the public testnet.

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

_Note: The default Hardhat tasks are still visible (e.g `compile`), many of these will do nothing as we are only using
Hardhat for tasks/testing._

Help information for a specific task can be obtained using

```
npx hardhat help <TASK>
```

e.g:

```
npx hardhat help validatorPool:submitApplication

