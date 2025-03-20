# mezo perf tests

## Some values being modified accross runs

`max gas per block`: (.consensus.params.block.max_gas in the genesis file), this is set by default
to 10_000_000 right now.

`mempool size`: set to 10k for now, could be set to 20k.

These two parameter should be set / configured quite tighly as increasing the number of transactions in
the mempool could be problematic if the max gas per block is not big enough to execute transaction quickly.

All these really need to be set according to the minimum recommended hardware to the validators.

## Install

To build / install the scripts run the following command inside the folder:

```
go install # install globaly
go build   # build the binary inside the current folder
```

## Using a test account

All commands comes with the following flags:

* -mnemonic: a bip39 mnemonic as a string
* -localkey: a path to the seed account of a local testnet
* -privkey: an hex encode private key

These are use to specify which dev account to use to run the perf tests. Usually when using a node running on the same
machine, the following should work:

```
perfs <command> -localkey=<PATH_TO_HOME>/.localnode/dev0_key_seed.json
```

## Deploy an erc20 token for the test

The following command will deploy a new erc20 token (base ERC20 from open zeppelin), and mint 10000 of it, sent to the
dev deployer. The deployer account is the one specified by via the localkey / mnemonic or privkey.

```
perfs deploy_token -localkey "../mezod/.localnode/dev0_key_seed.json"
```

The address of the deployed token is returned, take not of this for later.

## Generate accounts and transfer funds

The following commands takes a count argument which is the number of accounts to generate / topup

### Generate and topup account with gas/native funds

```
perfs generate -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json
```

### Topup with native funds

```
perfs topup -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json
```

### Topup with erc20 funds

```
perfs topup_erc20 -address=<TOKEN_ADDRESS> -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json
```

## Run tests

The following commands takes a count argument which is the number of accounts to use to send transaction.
They also allow for the account to wait for the transaction receipt or not, if we do not wait for the receipt,
a small sleep is used in order to leave time to process transactions and avoid having nonce issues.

All these following commands

### Run native transfer

This execute a native transfer (transfering value as part of the transation), and send a small amount from the address
 to another.

```
perfs run_native -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json
```

### Run ERC20 precompile transfer

This execute a native transfer via the erc20 precompile.

```
perfs run_erc20_precompile -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json
```

### Run ERC20 transfer

This execute actual ERC20 transfer, the token needs to be deployed, and topup_erc20 ran before.

```
perfs run_erc20 -address=<TOKEN_ADDRESS> -count=<NUMBER_OF_ADDRESSES> -localkey=../mezod/.localnode/dev0_key_seed.json


## Aggregate results
