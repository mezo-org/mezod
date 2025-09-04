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

### Setup Google OAuth login

First, create an OAuth application in Google Cloud Console:
- Go to Google Cloud Console
- Navigate to APIs & Services > Credentials
- Click Create Credentials > OAuth 2.0 Client IDs
- Set Application type to Web application
- Add your authorized redirect URIs: `https://<GRAFANA_DOMAIN>/grafana/login/google`

Save and note down the Client ID and Client Secret.

### Secrets

#### Grafana

Secrets are used to set the default user and password to log in to
the Grafana UI. Here are the required entries under grafana-secret:
- admin-password
- admin-user

Use `kubectl` to create the secrets or apply changes:
```Shell
kubectl create secret generic -n monitoring grafana-secret \
  --from-literal=admin-user=<USER> \
  --from-literal=admin-password=<PASSWORD>

```

And under grafana-auth-google-secret:
- client-id
- secret-id

Use `kubectl` to create the secrets or apply changes:
```Shell
kubectl create secret generic -n monitoring grafana-auth-google-secret \
  --from-literal=client-id=<CLIENT_ID> \
  --from-literal=client-secret=<CLIENT_SECRET>

```

#### Metrics Scraper

One secret is used to set the configuration of the metrics scraper service. To
create it, use the following command:
```Shell
kubectl create secret generic metrics-scraper-config -n monitoring \
  --from-literal=chain-id=<CHAIN_ID> \
  --from-file=nodes-config.json=<PATH_TO_NODES_CONFIG>
```

Here's an example of the `nodes-config.json` configuration file:
```
{
  "nodes": [
    {
      "rpc_url": "http://<MEZO_RPC_URL>:8545",
      "moniker": "<MEZO_NODE_MONIKER>"
    }
  ]
}
```

### Static IP for the metrics-scraper service

The metrics scraper service requires a static IP which is to be allowlisted
by node operators so the service can access them. It is the IP of the Cloud NAT,
which is created as part of the `mezo-<environment>` Terraform configuration.

Node operators will need to allowlist this IP on their EVM JSON-RPC port
(the default being 8545).

### Ingresses

`monitoring` defines one ingress resource (ingress.yaml), which  creates an external
access point for Grafana. This allows external users to securely access the Grafana
monitoring dashboard through HTTPS at the path "/grafana". It is pinned to the
`mezo-<environment>-monitoring-hub-external-ip` static global IP and uses
`mezo-<environment>-monitoring-hub-ssl-certificate` SSL certificate, both created by
the `mezo-<environment>` Terraform module.

### Add a node to the monitoring

The monitoring system does not automatically discover new nodes yet. That said,
new nodes must be added to the monitoring system manually. This can be done by
running the `add-node.sh` script on the desired environment.

First, switch to the desired environment by setting the right `kubectl` context:
```Shell
kubectl config use-context <context-name>
```

You can use `kubectl config get-contexts` to list the available contexts. If you don't
see the context you need, you have to download cluster credentials using the
`gcloud container clusters get-credentials` command.

Then, run the `add-node.sh` script:
```Shell
./add-node.sh <rpc_url> <moniker>
```
where `<rpc_url>` is the EVM JSON-RPC URL of the node and `<moniker>` is the moniker that
will be displayed in the monitoring dashboard.
