# Kubernetes: monitoring

This module contains Kubernetes deployments for monitoring the Mezo nodes.

### Prerequisites

- Infrastructure components for the `mezo-<environment>` GCP project created using the
  corresponding Terraform module (make sure GKE version is >=1.29)
- `gcloud` installed and authorized to access the `mezo-<environment>` GCP project
- `kubectl` tool installed

### Monitoring namespace

All the kubernetes monitoring infrastructure lives inside the `monitoring`
namespace. Before starting, you need to create the namespace using the
following command:
```Shell
kubectl create namespace monitoring
```

### Secrets

#### Grafana

Secrets are used to set the default user and password to log in to
the Grafana UI. Here are the required entries:
- admin-password
- admin-user

Use `kubectl` to create the secrets or apply changes:
```Shell
kubectl create secret generic -n monitoring grafana-secret \
  --from-literal=admin-user=<USER> \
  --from-literal=admin-password=<PASSWORD>
```

#### Metrics Scraper

One secret is used to set the configuration of the metrics scraper service. To
create it, use the following command:
```Shell
kubectl create secret generic metrics-scraper-config -n monitoring \
  --from-file=config.json=<PATH_TO_CONFIG>
```

Here's an example of the configuration:
```
{
    "poll_rate": "2s",
    "chain_id": "mezo_31611-1",
    "nodes": [
	{
	    "rpc_url": "http://<MEZO_RPC_URL>:8545",
	    "moniker": "<MEZO_NODE_MONIKER>"
	},
	...
	]
}

```

Moreover you can find the Go struct definition of the config in the following
file: https://github.com/mezo-org/mezod/blob/main/metrics-scraper/config.go

### Static IP for the metrics-scraper service

The metrics scraper service requires a static IP which is to be allowlisted
by node operators so the service can access them. It is pinned to
`mezo-<environment>-monitoring-external-ip`, which is created as part of the
mezo-<environment> Terraform configuration.

Node operators will need to allowlist this IP on their EVM JSON-RPC port
(the default being 8545).

### Ingresses

`monitoring` defines one ingress resource (ingress.yaml), which  creates an external
access point for Grafana. This allows external users to securely access the Grafana
monitoring dashboard through HTTPS at the path "/grafana". It is pinned to the
`mezo-<environment>-monitoring-hub-external-ip` static global IP and uses
`mezo-<environment>-monitoring-hub-ssl-certificate` SSL certificate, both created by
the `mezo-<environment>` Terraform module.
