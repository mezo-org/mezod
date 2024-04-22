# Terraform config.
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.25.0"
    }
  }
}

variable "project_id" {
  description = "Project ID"
}

# Provider.
provider "google" {
  project = var.project_id
}

# Terraform backend storage bucket.
resource "google_storage_bucket" "terraform_backend" {
  name     = "mezo-staging-terraform-backend-bucket"
  location = "US"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}