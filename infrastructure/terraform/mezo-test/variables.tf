variable "services" {
  type = list(string)
  description = "Service APIs used in the project"

  default = [
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
  ]
}

variable "project_name" {
  description = "Project name"
  default = "mezo-test"
}

variable "project_id" {
  description = "Project ID"
}

variable "region" {
  type        = object({name = string, zones = list(string)})
  description = "Region and zones info"

  default = {
    name  = "us-central1"
    zones = ["us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"]
  }
}

variable "vpc_network" {
  type        = map(string)
  description = "VPC network data"

  default = {
    name = "mezo-test-vpc-network"
  }
}

variable "gke_subnet" {
  type        = map(string)
  description = "Subnet for deploying GKE cluster resources"

  default = {
    name             = "mezo-test-gke-subnet"
    primary_ip_range = "10.1.0.0/16"

    pods_ip_range_name = "mezo-test-gke-pods-ip-range"
    pods_ip_range      = "10.100.0.0/16"

    services_ip_range_name = "mezo-test-gke-services-ip-range"
    services_ip_range      = "10.101.0.0/16"
  }
}