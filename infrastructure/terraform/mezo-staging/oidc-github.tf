#
# GitHub Actions OIDC configuration
#
# References:
# https://cloud.google.com/blog/products/identity-security/enabling-keyless-authentication-from-github-actions
# https://github.com/terraform-google-modules/terraform-google-github-actions-runners/blob/master/modules/gh-oidc/README.md
#

resource "google_iam_workload_identity_pool" "github_pool" {
  workload_identity_pool_id = "github-pool"
  display_name              = "GitHub pool"
  description               = "Identity pool for GitHub deployments"
}

resource "google_iam_workload_identity_pool_provider" "github_provider" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github_pool.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-provider"
  display_name                       = "GitHub provider"
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }
  attribute_condition = "assertion.repository_owner==\"${var.oidc_github.github_organization}\""


  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

#
# Service Account for GitHub Actions
#
resource "google_service_account" "mezo_staging_gha" {
  account_id   = var.oidc_github.service_account
  display_name = "GitHub Actions Service Account"
  description  = "Service Account for Github OIDC to push images to Artifact Registry"
}

resource "google_artifact_registry_repository_iam_member" "mezo_staging_gha" {
  repository = google_artifact_registry_repository.docker_public.name
  role       = "roles/artifactregistry.writer"
  member     = "serviceAccount:${google_service_account.mezo_staging_gha.email}"
}

# Attach the Workload Identity Pool (via role) to the Service Accounts
# that will be used by GitHub Actions.
resource "google_service_account_iam_member" "mezo_staging_gha" {
  service_account_id = google_service_account.mezo_staging_gha.email
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github_pool.name}/attribute.repository/${var.oidc_github.github_organization}/${var.oidc_github.repository}"
}
