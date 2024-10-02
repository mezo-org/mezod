variable "services" {
  type = list(string)
  description = "Service APIs used in the project"

  default = [
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
  ]
}

variable "project_id" {
  description = "Project ID"
  type = string
}

variable "region" {
  type        = object({name = string, zones = list(string)})
  description = "Region and zones info"

  default = {
    name  = "us-central1"
    zones = ["us-central1-a", "us-central1-b", "us-central1-c"]
  }
}

variable "vpc_network" {
  type        = map(string)
  description = "VPC network data"

  default = {
    name = "mezo-staging-vpc-network"
  }
}

variable "gke_subnet" {
  type        = map(string)
  description = "Subnet for deploying GKE cluster resources"

  default = {
    name             = "mezo-staging-gke-subnet"
    primary_ip_range = "10.1.0.0/16"

    pods_ip_range_name = "mezo-staging-gke-pods-ip-range"
    pods_ip_range      = "10.100.0.0/16"

    services_ip_range_name = "mezo-staging-gke-services-ip-range"
    services_ip_range      = "10.101.0.0/16"
  }
}

variable "cloud_nat" {
  type        = map(string)
  description = "Cloud NAT info"

  default = {
    name        = "mezo-staging-cloud-nat"
    router_name = "mezo-staging-cloud-router"
  }
}

variable "gke_cluster" {
  type        = object({
    name                   = string
    node_pool_name         = string
    node_pool_machine_type = string
    node_pool_size         = number
  })
  description = "GKE cluster info"

  default = {
    name                   = "mezo-staging-gke-cluster"
    node_pool_name         = "mezo-staging-gke-node-pool"
    node_pool_machine_type = "n2-standard-8"
    node_pool_size         = 1
  }
}

variable "external_ip_addresses" {
  type = list(string)
  description = "External IP addresses reserved for the project"

  default = [
    "mezo-staging-node-0-external-ip",
    "mezo-staging-node-1-external-ip",
    "mezo-staging-node-2-external-ip",
    "mezo-staging-node-3-external-ip",
    "mezo-staging-node-4-external-ip",
  ]
}

variable "global_external_ip_addresses" {
  type = list(string)
  description = "Global external IP addresses reserved for the project"

  default = [
    "mezo-staging-blockscout-api-external-ip",
    "mezo-staging-blockscout-app-external-ip",
  ]
}

variable "gcf" {
  type        = map(string)
  description = "Cloud functions info"

  default = {
    archive_bucket_name = "mezo-staging-gcf-archive-bucket"
    faucet_function_name = "mezo-staging-faucet-function"
  }
}