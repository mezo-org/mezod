module "gke" {
  source     = "terraform-google-modules/kubernetes-engine/google//modules/private-cluster"
  version    = "36.2.0"
  depends_on = [module.vpc]

  project_id                   = var.project_id
  name                         = var.gke_cluster.name
  region                       = var.region.name
  regional                     = true
  zones                        = var.region.zones
  network                      = var.vpc_network.name
  subnetwork                   = var.gke_subnet.name
  ip_range_pods                = var.gke_subnet.pods_ip_range_name
  ip_range_services            = var.gke_subnet.services_ip_range_name
  remove_default_node_pool     = true
  enable_private_nodes         = true
  deletion_protection          = false

  node_pools = [
    {
      name         = var.gke_cluster.node_pool_name
      machine_type = var.gke_cluster.node_pool_machine_type
      autoscaling  = false
      node_count   = var.gke_cluster.node_pool_size
    }
  ]
}

resource "google_storage_bucket" "explorer_logos_bucket" {
  name     = "mezo-production-explorer-logos-bucket"
  location = var.region.name

  uniform_bucket_level_access = true
  force_destroy               = false
}

resource "google_storage_bucket_iam_member" "public_read" {
  bucket = google_storage_bucket.explorer_logos_bucket.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

output "explorer_logos_bucket_url" {
  value = "https://storage.googleapis.com/${google_storage_bucket.explorer_logos_bucket.name}"
}
