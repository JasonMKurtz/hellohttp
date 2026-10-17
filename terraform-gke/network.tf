# A VPC with one subnet. Unlike AWS, GCP VPCs are global and subnets are
# regional (not per-zone), so one subnet covers all three zones.
#
# Nodes are private (no public IPs) and there's no Cloud NAT, so this costs
# nothing while idle. Private Google Access lets them pull images from
# Artifact Registry anyway. Incoming traffic arrives through the load
# balancer, which needs no route out.

resource "google_compute_network" "this" {
  name                    = var.cluster_name
  auto_create_subnetworks = false

  depends_on = [google_project_service.this]
}

resource "google_compute_subnetwork" "this" {
  name                     = var.cluster_name
  network                  = google_compute_network.this.id
  region                   = var.region
  ip_cidr_range            = var.subnet_cidr
  private_ip_google_access = true

  # GKE is VPC-native: every pod and Service gets a real VPC IP from these
  # ranges, so the load balancer can send traffic straight to pods.
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.4.0.0/14"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.8.0.0/20"
  }
}
