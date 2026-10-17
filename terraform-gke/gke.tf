# GKE Standard: you choose and pay for the nodes (VMs), and Google runs the
# control plane. Autopilot would hide the nodes and bill per pod instead,
# which has no free tier.

locals {
  cluster_count = var.cluster_enabled ? 1 : 0
  node_tag      = "${var.cluster_name}-node"
}

resource "google_container_cluster" "this" {
  count = local.cluster_count

  name     = var.cluster_name
  location = var.zone

  network    = google_compute_network.this.id
  subnetwork = google_compute_subnetwork.this.id

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # A cluster can't be created without nodes, so GKE makes a default node
  # pool, which Terraform deletes right away in favor of the one below. That
  # way node settings can change without recreating the cluster.
  remove_default_node_pool = true
  initial_node_count       = 1

  # Google picks and upgrades the Kubernetes version from this channel.
  release_channel {
    channel = "REGULAR"
  }

  # The Cloud Logging and Monitoring agents run on every node and need more
  # memory than an e2-micro can spare. `kubectl logs` still works without
  # them; logs just aren't kept or shown in the console.
  logging_service    = "none"
  monitoring_service = "none"

  # On by default in the provider; it would block cluster_enabled=false.
  deletion_protection = false
}

resource "google_container_node_pool" "default" {
  count = local.cluster_count

  name       = "default"
  cluster    = google_container_cluster.this[0].id
  location   = var.zone
  node_count = 1

  node_config {
    machine_type = var.node_machine_type
    spot         = var.node_spot

    # The free tier includes 30 GB of standard (HDD) disk. GKE's default is
    # 100 GB of balanced (SSD) disk.
    disk_type    = "pd-standard"
    disk_size_gb = 30

    service_account = google_service_account.node.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    tags = [local.node_tag]
  }

  depends_on = [
    google_project_iam_member.node,
    google_artifact_registry_repository_iam_member.node,
  ]
}
