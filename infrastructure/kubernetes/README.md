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
gcloud container clusters get-credentials --project <google_cloud_project_id> mezo-<environment>-gke-cluster --region=us-central1 --dns-endpoint
```

The `--dns-endpoint` flag is required. The control plane of each cluster
answers only on its DNS-based endpoint, where Google IAM checks the caller
before the request reaches the API server; the public IP endpoint is
disabled by the Terraform module. A kubeconfig entry created without the
flag points at the private IP of the control plane, which is reachable
only from inside the VPC.

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
