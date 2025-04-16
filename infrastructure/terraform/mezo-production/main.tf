terraform {
  backend "gcs" {
    bucket = "mezo-production-terraform-backend-bucket"
    prefix = "terraform/state"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.30.0"
    }
  }

  required_version = "~> 1.9.6"
}

# Service Usage API must be enabled manually to use this resource:
# https://console.cloud.google.com/apis/api/serviceusage.googleapis.com
resource "google_project_service" "services" {
  for_each                   = toset(var.services)
  service                    = each.value
  disable_dependent_services = true
}
