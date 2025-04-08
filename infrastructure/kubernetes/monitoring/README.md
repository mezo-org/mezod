# Mezo monitoring

## Secrets

Grafana requires an admin user and password, you can create those by setting the following
kubernetes secrets:

```Shell
kubectl create secret generic my-secret \
  --from-literal=username=admin \
  --from-literal=password=super-secret-password
```

TODO: more steps
