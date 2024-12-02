resource "google_compute_ssl_certificate" "mezo_staging_explorer" {
  name = "mezo-staging-explorer-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-staging-explorer.key")
  certificate = file("./ssl-certificates/mezo-staging-explorer.crt")

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_ssl_certificate" "mezo_staging_rpc" {
  name = "mezo-staging-rpc-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-staging-rpc.key")
  certificate = file("./ssl-certificates/mezo-staging-rpc.crt")

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_ssl_certificate" "mezo_staging_rpc-ws" {
  name = "mezo-staging-rpc-ws-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-staging-rpc-ws.key")
  certificate = file("./ssl-certificates/mezo-staging-rpc-ws.crt")

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_ssl_certificate" "mezo_staging_safe" {
  name = "mezo-staging-safe-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-staging-safe.key")
  certificate = file("./ssl-certificates/mezo-staging-safe.crt")

  lifecycle {
    create_before_destroy = true
  }
}