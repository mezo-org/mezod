# Mezo monitoring

## Secrets

Grafana requires an admin user and password, you can create those by setting the following
kubernetes secrets:

```Shell
kubectl create secret generic -n monitoring grafana-secret \
  --from-literal=staging-admin-user=<STAGING_USER> \
  --from-literal=staging-admin-password=<STAGING_PASSWORD> \
  --from-literal=production-admin-user=<PRODUCTION_USER> \
  --from-literal=production-admin-password=<PRODUCTION_PASSWORD>
```

TODO: more steps
