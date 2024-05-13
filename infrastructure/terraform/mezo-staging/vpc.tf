module "vpc" {
  source     = "terraform-google-modules/network/google"
  version    = "9.1.0"
  depends_on = [google_project_service.services]

  project_id   = var.project_id
  network_name = var.vpc_network.name

  subnets = [
    {
      subnet_name   = var.gke_subnet.name
      subnet_ip     = var.gke_subnet.primary_ip_range
      subnet_region = var.region.name
    }
  ]

  secondary_ranges = {
    (var.gke_subnet.name) = [
      {
        range_name    = var.gke_subnet.pods_ip_range_name
        ip_cidr_range = var.gke_subnet.pods_ip_range
      },
      {
        range_name    = var.gke_subnet.services_ip_range_name
        ip_cidr_range = var.gke_subnet.services_ip_range
      }
    ]
  }
}

resource "google_compute_address" "external_ip_addresses" {
  for_each = toset(var.external_ip_addresses)
  name = each.key
}

resource "google_compute_global_address" "global_external_ip_addresses" {
  for_each = toset(var.global_external_ip_addresses)
  name = each.key
}