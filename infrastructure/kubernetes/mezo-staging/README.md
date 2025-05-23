# Kubernetes: mezo-staging

This module contains Kubernetes deployments for the `mezo-staging-gke-cluster` cluster
created by the corresponding [Terraform module](./../../terraform/mezo-staging/README.md).

## General guide

Please refer to the [general Kubernetes package guide](../README.md) for instructions.

## Helm releases

This module defines the following Helm releases.

### Mezo validators

This Helm release defines 5 Mezo validators for the testnet chain.
The release consists of the following resources.

#### Secrets

Sensitive testnet artifacts like private keys are ported to the Kubernetes
cluster as secrets: `mezo-node-<index>`. Each validator requires a separate
secret containing the following keys:
- `ETHEREUM_ENDPOINT`
- `KEYRING_MNEMONIC`
- `KEYRING_NAME`
- `KEYRING_PASSWORD`

These keys are used in the environment variables of the Mezo node. They must
be created before deploying the Mezo validators. Use `kubectl` to create the
secrets or apply changes (example for `mezo-node-0`):
```shell
kubectl create secret generic mezo-node-0 \
  -n default \
  --from-literal=ETHEREUM_ENDPOINT="wss://<provide_ethereum_endpoint>" \
  --from-literal=KEYRING_NAME=mezo-node-0 \
  --from-literal=KEYRING_PASSWORD="<provide_password>" \
  --from-literal=KEYRING_MNEMONIC="<provide_mnemonic>"
```

#### Stateful sets

Initial Mezo validators are defined as Kubernetes `mezo-node-<index>` stateful sets using Helmfile.
They mount the aforementioned secret. Moreover, each validator uses a persistent volume to hold their 
validator state.

To deploy a single Mezo validator using Helmfile, run the following command:
```shell
helmfile apply -i -l name=mezo-node-<index>
```

All Mezo validators are exposed publicly using external load balancers
pinned to `mezo-staging-node-<index>-external-ip` static IPs created by
the `mezo-staging` Terraform module.

### Blockscout explorer

This Helm release defines a Blockscout explorer for the testnet chain.
The release creates two stateful sets, for the app and the API, and
exposes them using distinct ingress resources.

Before applying the Blockscout stack, create the following secret
(only if it does not exist yet):

```shell
kubectl create secret generic blockscout-stack -n default \
  --from-literal=WALLET_CONNECT_PROJECT_ID=<wallet-connect-project-id>
```

### Postgres database

This Helm release defines a Postgres database that is used as
a persistent storage for the Blockscout explorer.

## Standalone ingresses

`mezo-staging` defines two standalone ingress resources:
- `mezo-rpc` exposing the Mezo JSON-RPC API over HTTP. It is pinned to the
  `mezo-staging-rpc-external-ip` static global IP and uses `mezo-staging-rpc-ssl-certificate`
  SSL certificate, both created by the `mezo-staging` Terraform module.
  This ingress hits the 8545 port of the `mezo-rpc` Kubernetes service which
  load balances the requests to the Mezo nodes deployed on the cluster.
- `mezo-rpc-ws` exposing the Mezo JSON-RPC API over WebSockets. It is pinned
  to the `mezo-staging-rpc-ws-external-ip` static global IP and uses
  `mezo-staging-rpc-ws-ssl-certificate` SSL certificate, both created by the
  `mezo-staging` Terraform module. This ingress hits the 8546 port of the `mezo-rpc`
  Kubernetes service which load balances the requests to the Mezo nodes deployed
  on the cluster. The WebSocket endpoint is deployed under a separate domain,
  not as a custom path of the HTTP endpoint due to the limitations of the
  default GCP ingress controller (gce). A custom path would have to be rewritten
  to `/` being the base path of the JSON-RPC API. However, the default
  GCP ingress controller does not support path rewriting.

