# Kubernetes

This package contains Kubernetes configuration for the `mezo-<environment>` clusters.
Each sub-directory contains Kubernetes configuration for a single cluster representing
the given Mezo environment. This general guide applies to all environments. Please consult
README files in each sub-directory for environment-specific instructions.

## Prerequisites

- Infrastructure components for the `mezo-<environment>` GCP project created using the
  corresponding Terraform module (make sure GKE version is >=1.29)
- `gcloud` installed and authorized to access the `mezo-<environment>` GCP project
- `kubectl` tool installed
- `helm` tool installed
- `helmfile` tool installed (run `helmfile init` to install required Helm plugins)

## Authentication

Use `gcloud` to get credentials for the cluster and automatically
configure `kubectl`:

```shell
gcloud container clusters get-credentials mezo-<environment>-gke-cluster --region=us-central1
```

Verify that everything went as expected and `kubectl` points to the correct cluster:
```shell
kubectl config current-context
```

## Resource management

The main tool to manage resources in the clusters is `helmfile`.

To apply changes to the resources, run the following command:
```shell
helmfile apply
```

To check the differences between the current state and the desired state, 
run:
```shell
helmfile diff
```
