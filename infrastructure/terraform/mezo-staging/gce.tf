resource "google_compute_instance" "safe_infrastructure" {
  name         = "mezo-staging-safe-infrastructure"
  machine_type = "n2-standard-4"
  zone         = var.region.zones[0]

  tags = ["mezo-staging-safe-infrastructure"]

  boot_disk {
    auto_delete = false

    initialize_params {
      size = 100
      type = "pd-ssd"
      image = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-2404-noble-amd64-v20241115"
    }
  }

  network_interface {
    network    = module.vpc.network_name
    subnetwork = var.gce_subnet.name

    # Deliberately omit the access_config block to make sure the instance does
    # not have a public IP.
  }

  allow_stopping_for_update = true
}

