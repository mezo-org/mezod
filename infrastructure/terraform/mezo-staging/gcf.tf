resource "google_storage_bucket" "gcf_archive" {
  name                        = var.gcf.archive_bucket_name
  location                    = "US"
  uniform_bucket_level_access = true
  force_destroy = true
  versioning {
    enabled = true
  }
}

data "archive_file" "faucet_function" {
  type        = "zip"
  source_dir  = "gcf/faucet/"
  output_path = "archive/${var.gcf.faucet_function_name}.zip"
}

resource "google_storage_bucket_object" "faucet_function" {
  name   = "${var.gcf.faucet_function_name}.zip"
  bucket = google_storage_bucket.gcf_archive.name
  source = data.archive_file.faucet_function.output_path
}

resource "google_cloudfunctions2_function" "faucet" {
  name        = var.gcf.faucet_function_name
  location    = var.region.name
  description = "the faucet's distribute function"

  build_config {
    runtime     = "go122"
    entry_point = "Distribute"
    source {
      storage_source {
        bucket = google_storage_bucket.gcf_archive.name
        object = google_storage_bucket_object.faucet_function.name
      }
    }
  }

  service_config {
    max_instance_count = 1
    available_memory   = "256M"
    timeout_seconds    = 60
    environment_variables = {
      RPC_URL     = local.faucet_config.rpc_url
      PRIVATE_KEY = local.faucet_config.private_key
    }
  }
}

resource "google_cloud_run_service_iam_member" "member" {
  location = google_cloudfunctions2_function.faucet.location
  service  = google_cloudfunctions2_function.faucet.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

output "function_uri" {
  value = google_cloudfunctions2_function.faucet.service_config[0].uri
}

locals {
  faucet_config = sensitive(yamldecode(file("./configs/faucet-config.yaml")))
}