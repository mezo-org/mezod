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
