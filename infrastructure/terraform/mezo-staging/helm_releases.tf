resource "helm_release" "postgresql" {
  depends_on = [module.gke]
  name       = "postgresql"
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "postgresql"
  version    = "15.2.10"

  values = [
    file("../../helm/mezo-staging/postgresql-values.yaml")
  ]
}

resource "helm_release" "redis" {
  depends_on = [module.gke]
  name       = "redis"
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "redis"
  version    = "19.3.0"

  values = [
    file("../../helm/mezo-staging/redis-values.yaml")
  ]
}