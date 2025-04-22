# Kubernetes: mezo-staging/monitoring

This module contains Kubernetes deployments for monitoring the mezo staging nodes.

### Prerequisites

- Infrastructure components for the `mezo-<environment>` GCP project created using the
  corresponding Terraform module (make sure GKE version is >=1.29)
- `gcloud` installed and authorized to access the `mezo-<environment>` GCP project
- `kubectl` tool installed

### Secrets

#### Grafana

Secrets are used to set the default user and password to log in to
the Grafana UI. Here are the required entries:
- admin-password
- admin-user

Use `kubectl` to create the secrets or apply changes:
```Shell
kubectl create secret generic -n monitoring grafana-secret \
  --from-literal=admin-user=<STAGING_USER> \
  --from-literal=admin-password=<STAGING_PASSWORD>
```

#### Metrics Scraper

One secret is used to set the configuration of the metrics scraper service. To
create it, use the following command:
```Shell
kubectl create secret generic metrics-scraper-config -n monitoring \
  --from-file=config.json=<PATH_TO_CONFIG>
```

### Static IP for the metrics-scraper service

The metrics scraper service requires a static IP which is to be allowlisted
by node operators so the service can access them. It is pinned to
`mezo-<environment>-monitoring-external-ip`, which is created as part of the
mezo-staging Terraform configuration.

### Ingresses

`monitoring` defines one ingress resource (ingress.yaml). It is pinned to the
  `mezo-<environment>-monitoring-hub-external-ip` static global IP and uses
  `mezo-<environment>-monitoring-hub-ssl-certificate` SSL certificate, both created by
  the `mezo-<environment>` Terraform module.