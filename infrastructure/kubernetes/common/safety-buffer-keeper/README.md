# Safety Buffer Keeper Deployment

This directory contains Kubernetes manifests for deploying the safety buffer keeper service.

## Configuration

The deployment expects a secret named `safety-buffer-keeper` with:

- `private-key`: Value injected into `SAFETY_BUFFER_KEEPER_PRIVATE_KEY`
- `rpc-url`: Value injected into `MEZO_BOAR_HTTPS_URL`

## Files

- `deployment.yaml`: Main Kubernetes deployment
- `kustomization.yaml`: Kustomize configuration for this component

## Usage

1. Create or update the required secret:

```shell
kubectl create secret generic safety-buffer-keeper \
  --from-literal=private-key="<private_key>" \
  --from-literal=rpc-url="<rpc_url>" \
  --namespace=default \
  --dry-run=client -o yaml | kubectl apply -f -
```

2. Apply manifests from the target environment overlay:

```shell
kubectl apply -k .
```

3. Restart deployment when needed:

```shell
kubectl rollout restart deployment/safety-buffer-keeper -n default
```
