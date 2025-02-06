# System tests

This project demonstrates system tests for BTC transfers on Mezo chain.

## Run Mezo localnet

Clean and init localnet binaries

```shell
make localnet-bin-clean && make localnet-bin-init
```

Start 3 or 4 nodes

```shell
make localnet-bin-start
```

## Run hardhat system tests against localnet

```shell
npm run test
```
