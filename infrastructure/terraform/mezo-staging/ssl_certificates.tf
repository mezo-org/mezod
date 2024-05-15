resource "google_compute_ssl_certificate" "mezo_staging_explorer" {
  name = "mezo-staging-explorer-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-staging-explorer.key")
  certificate = file("./ssl-certificates/mezo-staging-explorer.crt")

  lifecycle {
    create_before_destroy = true
  }
}