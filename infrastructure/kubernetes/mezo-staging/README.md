# Kubernetes: mezo-staging

This module contains Kubernetes deployments for the `mezo-staging-gke-cluster` cluster
created by the corresponding [Terraform module](./../../terraform/mezo-staging/README.md).

### Prerequisites

- Infrastructure components for the `mezo-staging` GCP project created using the
  corresponding Terraform module (make sure GKE version is >=1.29)
- `gcloud` installed and authorized to access the `mezo-staging` GCP project
- `kubectl` tool installed
- If `generate-mezo-node-keystore.sh` script is used, testnet artifacts
  must be present in the `.public-testnet` directory (see `scripts/public-testnet.sh`)

### Authentication

Use `gcloud` to get credentials for the cluster and automatically 
configure `kubectl`:

```shell
gcloud container clusters get-credentials mezo-staging-gke-cluster --region=us-central1
```

Verify that everything went as expected and `kubectl` points to the correct cluster:
```shell 
kubectl config current-context
```

### Config maps

Testnet artifacts are ported to the Kubernetes cluster as config maps:
- `mezo-node-config`: Config map holding the three configuration files
  `app.toml`, `client.toml`, and `config.toml` consumed by Mezo validators. 
  This is a common config map re-used by all Mezo validators forming the initial set.
- `mezo-genesis-config`: Config map storing the genesis file necessary to 
  bootstrap the Mezo test chain. It is common for all Mezo validators.

Use `kubectl` to create the config maps or apply changes:
```shell
kubectl apply -f mezo-<name>-config.yaml
```

### Secrets

Sensitive testnet artifacts like private keys are ported to the Kubernetes
cluster as secrets:
- `mezo-node-<index>-keystore`: Secrets storing all private keys 
  (for different layers of the Cosmos SDK stack) used by Mezo validators. 
  Each secret is validator-specific. This secret is not defined directly as a k8s manifest. 
  Custom `generate-mezo-node-keystore.sh` script is used to inject those secrets 
  directly into the Kubernetes cluster.

### Stateful sets and services

Initial Mezo validators were defined as Kubernetes `mezo-node-<index>` stateful sets. 
They mount the aforementioned config maps and secrets as volumes and map their 
content to specific files expected by the Mezo node binary. Moreover, 
each validator uses a persistent volume to hold their validator state.

Each stateful set uses the `mezo-node` Docker image from the `mezo-staging-docker-internal`
Docker registry created by the `mezo-staging` Terraform module.

All Mezo validators are exposed publicly using external load balancers 
pinned to `mezo-staging-node-<index>-external-ip` static IPs created by 
the `mezo-staging` Terraform module.

Use `kubectl` to create the stateful sets (and services) or apply changes:
```shell
kubectl apply -f mezo-node-<index>.yaml
```