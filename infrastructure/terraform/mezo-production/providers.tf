provider "google" {
  project = var.project_id
  region  = var.region.name
  zone    = var.region.zones[0]
}
