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

The release runs the `mezo/blockscout` and `mezo/blockscout-frontend`
images from Docker Hub, pinned by digest in `values/blockscout-stack.yaml`.
Upstream publishes release images only to a private registry (backend
since 9.1.0, frontend since v2.4.0), so the images are built from the
upstream source with the Cloud Build configs in `infrastructure/cloudbuild`:

```shell
gcloud builds submit --no-source --project mezo-test-420708 \
  --config infrastructure/cloudbuild/blockscout-backend.yaml \
  --substitutions _VERSION=<backend-version>
gcloud builds submit --no-source --project mezo-test-420708 \
  --config infrastructure/cloudbuild/blockscout-frontend.yaml \
  --substitutions _VERSION=<frontend-tag>
```

The builds push to the `mezo-staging-docker-internal` Artifact Registry
repository. Pull the image from there, tag it as `mezo/blockscout:<version>`
or `mezo/blockscout-frontend:<tag>`, push it to Docker Hub and pin the
Docker Hub digest in the values file.

Backend v11.0.0 and later ship under the Blockscout Software Licence. It
allows a free explorer for your own chain and requires the "Made with
Blockscout" attribution, which the frontend footer shows by default.

To move to a new backend version, read the upstream release notes for the
upgrade path first. The backend refuses to start when the previous version
is too old: 11.x needs 10.2.3 or later with the
`UpdateInternalTransactionsPrimaryKey` migration completed, 12.x needs
11.0.2 or later with `FillInternalTransactionsAddressIds` completed. Check
the `migrations_status` table before the next hop. Scale the backend
deployment to zero before `helmfile apply`: the rolling update runs the old
and the new pod side by side and the two connection pools exceed the
Postgres `max_connections`.

The GCE ingress controller copies the readiness probe path into the load
balancer health check only when it creates the backend. The
`blockscout-api-backend-config.yaml` BackendConfig keeps the health check on
the current path. The release attaches it to the API Service in a postsync
hook.

Before applying the Blockscout stack, create the following secret
(only if it does not exist yet):

```shell
kubectl create secret generic blockscout-stack -n default \
  --from-literal=WALLET_CONNECT_PROJECT_ID=<wallet-connect-project-id> \
  --from-literal=SECRET_KEY_BASE=<secret-key-base>
```

Generate `SECRET_KEY_BASE` with `openssl rand -base64 48`. Blockscout
requires at least 64 characters.

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

