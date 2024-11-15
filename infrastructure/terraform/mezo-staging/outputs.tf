output "github_workload_identity_provider_id" {
  value = "${google_iam_workload_identity_pool.github_pool.name}/providers/${google_iam_workload_identity_pool_provider.github_provider.workload_identity_pool_provider_id}"
}
