# Kubernetes: mezo-production

This module contains Kubernetes deployments for the `mezo-production-gke-cluster` cluster
created by the corresponding [Terraform module](./../../terraform/mezo-production/README.md).

## General guide

Please refer to the [general Kubernetes package guide](../README.md) for instructions.

## Helm releases

This module defines the following Helm releases.

### Blockscout explorer

This Helm release defines a Blockscout explorer for the mainnet chain.
The release creates two stateful sets, for the app and the API, and
exposes them using distinct ingress resources.

The release runs the `mezo/blockscout` and `mezo/blockscout-frontend`
images from Docker Hub, pinned by digest in `values/blockscout-stack.yaml`.
The images are built from the upstream source with the Cloud Build configs
in `infrastructure/cloudbuild`; see the Blockscout section of the
`mezo-staging` README for the build and the upgrade notes. Rehearse every
backend version hop on staging first.

The GCE ingress controller copies the readiness probe path into the load
balancer health check only when it creates the backend. The
`blockscout-api-backend-config.yaml` BackendConfig keeps the health check on
the current path. The release attaches it to the API Service in a postsync
hook.

Before applying the Blockscout stack, create the following secret
(only if it does not exist yet):

```shell
kubectl create secret generic blockscout-stack -n default \
  --from-literal=ETHEREUM_JSONRPC_HTTP_URL=<mezo-rpc-http> \
  --from-literal=ETHEREUM_JSONRPC_WS_URL=<mezo-rpc-ws> \
  --from-literal=WALLET_CONNECT_PROJECT_ID=<wallet-connect-project-id> \
  --from-literal=SECRET_KEY_BASE=<secret-key-base>
```

Generate `SECRET_KEY_BASE` with `openssl rand -base64 48`. Blockscout
requires at least 64 characters.

### Postgres database

This Helm release defines a Postgres database that is used as
a persistent storage for the Blockscout explorer.

### Mezo private RPC node

The `mezo-rpc-node` Helm release is a private RPC node running against the mainnet chain.
This node's main purpose is debugging. Moreover, it can be also considered as a fallback
to support main RPC nodes provided by external partners if needed.

This release uses the `validator-kit` Helm chart but does not use the default
networking services provided by this chart. This is because the private RPC node
requires a custom port configuration (only p2p port should be public, the rest should be private)
and the `validator-kit` chart does not support this out of the box. To overcome this
problem, two services (one load balancer and one cluster IP) are defined manually
in the `mezo-rpc-node-service.yaml` file.
