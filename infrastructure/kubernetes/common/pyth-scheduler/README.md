# Pyth Scheduler (Price Pusher) Deployment

This directory contains Kubernetes deployment manifests for the Pyth Network price pusher service.

## Overview

The Pyth scheduler pushes price updates from the Pyth Network to the configured Mezo network.
It monitors price feeds and submits updates based on time and price deviation thresholds.

## Configuration

The deployment requires a secret named `pyth-scheduler-config` with the following keys:

- `endpoint`: RPC endpoint for Mezo network (e.g., "https://rpc.test.mezo.org")
- `pyth-contract-address`: Address of the Pyth contract on the EVM network (e.g., "0x2880aB155794e7179c9eE2e38200202908C17B43")
- `price-service-endpoint`: Hermes price service endpoint (e.g., "https://hermes.pyth.network")
- `mnemonic`: Mnemonic phrase for the wallet used to submit transactions

## Files

- `deployment.yaml`: Main Kubernetes deployment
- `kustomization.yaml`: Kustomize configuration for resource management
- `configmap.yaml`: ConfigMap for price feed configuration

## Usage

1. Create the required secret with your configuration
2. Run:

  ```Shell
  kubectl create secret generic pyth-scheduler-config \
    --from-literal=mnemonic="your twelve word mnemonic phrase here" \
    --from-literal=endpoint="<rpc_endpoint>" \
    --from-literal=pyth-contract-address="<pyth_contract_address>" \
    --from-literal=price-service-endpoint="https://hermes.pyth.network" \
    --namespace=default
  ```

3. Navigate to mezo-<environment> directory and apply the manifests: `kubectl apply -k .`

## Update configuration commands

```Shell
  kubectl delete configmap pyth-scheduler-price-config -n default
  kubectl apply -k .
  kubectl rollout restart deployment/pyth-scheduler -n default
```

## Add a new price feed

- Add a new price feed to the `configmap.yaml` file
- Find the price feed ID from the Pyth Feed IDs docs: https://docs.pyth.network/feeds/feed-ids
by copying the "Stable Price Feed ID" from the table
- Set the rest of the parameters such as `time_difference`, `price_deviation`, and `confidence_ratio`:
    - `time_difference`: Time difference in seconds between the price feed and the current time
    - `price_deviation`: Price deviation in percentage
    - `confidence_ratio`: Confidence ratio in percentage
- Run the following commands to apply the changes:

```Shell
  kubectl apply -k .
  kubectl rollout restart deployment/pyth-scheduler -n default
```
