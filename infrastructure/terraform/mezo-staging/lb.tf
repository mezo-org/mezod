module "load_balancer_safe" {
  source  = "terraform-google-modules/lb-http/google"
  version = "11.1.0"

  name    = "mezo-staging-safe"
  project = var.project_id

  create_address = false
  address        = google_compute_global_address.global_external_ip_addresses["mezo-staging-safe-external-ip"].address

  ssl              = true
  ssl_certificates = [google_compute_ssl_certificate.mezo_staging_safe.self_link]
  https_redirect   = true

  firewall_networks = [module.vpc.network_name]
  target_tags       = ["mezo-staging-safe-infrastructure"]

  backends = {
    default = {
      port       = "80"
      port_name  = "http"
      protocol   = "HTTP"
      enable_cdn = false

      health_check = {
        request_path = "/"
        port         = "80"
      }

      groups = [
        {
          group          = google_compute_network_endpoint_group.safe.id
          balancing_mode = "RATE"
          max_rate       = 100
        }
      ]

      iap_config = {
        enable = false
      }

      log_config = {
        enable = false
      }
    }
  }
}

resource "google_compute_network_endpoint_group" "safe" {
  name         = "mezo-staging-safe"
  network      = module.vpc.network_name
  subnetwork   = var.gce_subnet.name
  default_port = "80"
  zone         = var.region.zones[0]
}

resource "google_compute_network_endpoint" "safe" {
  network_endpoint_group = google_compute_network_endpoint_group.safe.name

  instance   = google_compute_instance.safe_infrastructure.name
  port       = google_compute_network_endpoint_group.safe.default_port
  ip_address = google_compute_instance.safe_infrastructure.network_interface[0].network_ip
}