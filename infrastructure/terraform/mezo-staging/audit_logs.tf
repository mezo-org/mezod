# GCP writes only Admin Activity logs by default, so a read of a secret, a
# bucket object or an IAM policy is not logged without this.
resource "google_project_iam_audit_config" "data_access" {
  for_each = toset([
    "compute.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com",
    "secretmanager.googleapis.com",
    "storage.googleapis.com",
  ])

  project = var.project_id
  service = each.value

  audit_log_config {
    log_type = "ADMIN_READ"
  }

  audit_log_config {
    log_type = "DATA_READ"
  }

  audit_log_config {
    log_type = "DATA_WRITE"
  }
}
