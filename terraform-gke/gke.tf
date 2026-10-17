# GKE Autopilot: Google manages nodes, scaling, and upgrades, and bills per
# pod (by its resource requests) instead of per node. The equivalent of EKS
# Auto Mode. Autopilot clusters are always regional: the control plane and
# pods are spread across the region's three zones.
resource "google_container_cluster" "this" {
  count = var.cluster_enabled ? 1 : 0

  name             = var.cluster_name
  location         = var.region
  enable_autopilot = true

  network    = google_compute_network.this.id
  subnetwork = google_compute_subnetwork.this.id

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # Nodes get no public IPs. The control plane keeps a public endpoint so
  # kubectl and Terraform can reach it from your laptop.
  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
  }

  # Google picks and upgrades the Kubernetes version from this channel.
  release_channel {
    channel = "REGULAR"
  }

  cluster_autoscaling {
    auto_provisioning_defaults {
      service_account = google_service_account.node.email
    }
  }

  # On by default in the provider; it would block cluster_enabled=false.
  deletion_protection = false

  depends_on = [
    google_project_iam_member.node,
    google_artifact_registry_repository_iam_member.node,
  ]
}
