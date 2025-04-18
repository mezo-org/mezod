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

### Postgres database

This Helm release defines a Postgres database that is used as
a persistent storage for the Blockscout explorer.
