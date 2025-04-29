resource "google_compute_ssl_certificate" "mezo_production_explorer" {
  name = "mezo-production-explorer-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-production-explorer.key")
  certificate = file("./ssl-certificates/mezo-production-explorer.crt")

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_ssl_certificate" "mezo_production_monitoring_hub" {
  name = "mezo-production-monitoring-hub-ssl-certificate"
  private_key = file("./ssl-certificates/mezo-production-monitoring-hub.key")
  certificate = file("./ssl-certificates/mezo-production-monitoring-hub.crt")

  lifecycle {
    create_before_destroy = true
  }
}
