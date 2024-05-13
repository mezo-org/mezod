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