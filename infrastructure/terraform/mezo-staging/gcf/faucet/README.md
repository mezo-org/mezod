# Faucet

The faucet distributes a portion of test BTC upon request. It has been implemented as a
[Google Cloud Function](https://cloud.google.com/functions/docs/writing).

## Run Locally

*It is recommended to use go version 1.22.x or greater.*

The faucet can be run locally with:

```
FUNCTION_TARGET=Distribute LOCAL_ONLY=true RPC_URL=http://127.0.0.1:8545 PRIVATE_KEY=<PRIVKEY> go run cmd/main.go
```

You can then make requests using your web browser by visiting

```
http://127.0.0.1:8080/ADDRESS
```

## Go bindings

The go bindings were generated using `abigen` (built from geth codebase) by doing the following:

```
cd $GOPATH/src/github.com/mezo-org
git clone https://github.com/mezo-org/go-ethereum.git
cd go-ethereum
make abigen
./build/bin/abigen \
--abi $GOPATH/src/github.com/mezo-org/mezod/precompile/btctoken/abi.json \
--out $GOPATH/src/github.com/mezo-org/mezod/infrastructure/terraform/mezo-staging/gcf/faucet/bindings/btctoken.go \
--pkg btctoken
```
