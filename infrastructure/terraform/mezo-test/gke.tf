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
  master_global_access_enabled = false
  deletion_protection          = false

  # Setting master_authorized_networks to an empty list does not disallow external
  # access to the control plane public endpoint. The authorized network
  # feature is disabled by default and the endpoint is accessible from the
  # internet. Using a non-empty list enables the authorized network feature
  # but also adds the specified networks to the allow-list. We want to restrict
  # external access completely and use a VPN to access the control plane.
  # This must be done beyond Terraform once the cluster is created.
  # Possible ways are:
  # - Use gcloud to update the cluster with the `--enable-master-authorized-networks`
  #   flag. Do not add any authorized networks using `--master-authorized-networks`.
  # - Use GCP console and set `Control plane authorized networks` to `Enabled`
  #   in the `Networking` section of the cluster settings. Do not add any
  #   authorized networks.
  #
  # For more details, consult:
  # https://cloud.google.com/kubernetes-engine/docs/how-to/authorized-networks
  master_authorized_networks = []

  node_pools = [
    {
      name         = var.gke_cluster.node_pool_name
      machine_type = var.gke_cluster.node_pool_machine_type
      autoscaling  = false
      node_count   = var.gke_cluster.node_pool_size
    }
  ]
}
