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