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

resource "helm_release" "blockscout_stack" {
  depends_on = [
    module.gke,
    helm_release.postgresql,
    google_compute_ssl_certificate.mezo_staging_explorer
  ]
  name       = "blockscout-stack"
  repository = "https://blockscout.github.io/helm-charts"
  chart      = "blockscout-stack"
  version    = "1.5.0"

  values = [
    file("../../helm/mezo-staging/blockscout-stack-values.yaml")
  ]
}