# Kubernetes: mezo-staging/monitoring

This module contains Kubernetes deployments for the monitoring of the mezo staging nodes.

### Prerequisites

- Infrastructure components for the `mezo-staging` GCP project created using the
  corresponding Terraform module (make sure GKE version is >=1.29)
- `gcloud` installed and authorized to access the `mezo-staging` GCP project
- `kubectl` tool installed

### Secrets

Secrets are use to set the default user and password to log in
the grafana UI. Here are the required entries:
- staging-admin-password
- staging-admin-user

Use `kubectl` to create the secrets or apply changes:
```Shell
kubectl create secret generic -n monitoring grafana-secret \
  --from-literal=staging-admin-user=<STAGING_USER> \
  --from-literal=staging-admin-password=<STAGING_PASSWORD>
```

### Static IP for the metrics-scraper service

The metrics scraper service requires a static IP which is to be allow listed
by node operator so the service can access them, it is pinned to
`mezo-staging-monitoring-external-ip` which is created as part of the
mezo-staging terraform configuration.

### Ingresses

`monitoring` defines one ingress resources:
- `mezo-rpc` exposing the Mezo JSON-RPC API over HTTP. It is pinned to the
  `mezo-staging-monitoring-hub-external-ip` static global IP and uses
  `mezo-staging-monitoring-hub-ssl-certificate` SSL certificate, both created by
  the `mezo-staging` Terraform module. This ingress hits the 3000 port of the
  `mezo-rpc` Kubernetes service which load balances the requests to the Mezo
  nodes deployed on the cluster.
