# Faucet

The faucet distributes 0.01 BTC upon request.

### Google Cloud Function

The faucet has been implemented as a [Google Cloud Function](https://cloud.google.com/functions/docs/writing). 

# Run Locally

The faucet can be run locally with:

```
FUNCTION_TARGET=Distribute LOCAL_ONLY=true RPCURL=http://127.0.0.1:8545 SECRET=<PRIVKEY> go run cmd/main.go
```

You can then make requests using your web browser by visiting

```
http://127.0.0.1:8080/ADDRESS
```

