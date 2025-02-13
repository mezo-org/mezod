# System tests

This project demonstrates system tests for the Mezo chain.

## Run Mezo single-node localnet

Start the node using the following command:

```shell
make localnode-bin-start
```

Wait a bit until the node's JSON-RPC server is up and running.

## Run hardhat system tests against localnet

```shell
./system-tests.sh
```
