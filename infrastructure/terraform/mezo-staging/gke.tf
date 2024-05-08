module "gke" {
  source     = "terraform-google-modules/kubernetes-engine/google//modules/private-cluster"
  version    = "30.2.0"
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

resource "google_artifact_registry_repository" "docker_internal" {
  location      = var.region.name
  repository_id = "mezo-staging-docker-internal"
  description   = "Docker repository for internal images"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository" "docker_public" {
  location      = var.region.name
  repository_id = "mezo-staging-docker-public"
  description   = "Docker repository for public images"
  format        = "DOCKER"
}

# Grant the GKE service account the ability to read from the internal Docker repository
resource "google_artifact_registry_repository_iam_member" "gke_sa-docker_internal_reader" {
  project = google_artifact_registry_repository.docker_internal.project
  location = google_artifact_registry_repository.docker_internal.location
  repository = google_artifact_registry_repository.docker_internal.name
  role = "roles/artifactregistry.reader"
  member = "serviceAccount:${module.gke.service_account}"
}

# Grant all users the ability to read from the public Docker repository
resource "google_artifact_registry_repository_iam_member" "all_users-docker_public_reader" {
  project = google_artifact_registry_repository.docker_public.project
  location = google_artifact_registry_repository.docker_public.location
  repository = google_artifact_registry_repository.docker_public.name
  role = "roles/artifactregistry.reader"
  member = "allUsers"
}
