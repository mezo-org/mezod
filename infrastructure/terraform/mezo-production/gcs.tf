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
