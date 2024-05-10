provider "google" {
  project = var.project_id
  region  = var.region.name
  zone    = var.region.zones[0]
}

data "google_client_config" "default" {}

provider "helm" {
  kubernetes {
    host                   = module.gke.endpoint
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(module.gke.ca_certificate)
  }
}