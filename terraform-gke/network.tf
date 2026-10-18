# A VPC with one subnet. Unlike AWS, GCP VPCs are global and subnets are
# regional (not per-zone).
#
# The node has a public IP, which is how you reach the app without a load
# balancer, and it reaches the internet through it. So there's no Cloud NAT.

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
  # ranges, so a load balancer can send traffic straight to pods.
  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.4.0.0/14"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.8.0.0/20"
  }
}

# GCP firewalls deny incoming traffic by default. Open the app's NodePort on
# the nodes (picked by network tag) to the internet. GKE adds its own rules
# for everything else, including load balancer health checks.
resource "google_compute_firewall" "node_port" {
  name    = "${var.cluster_name}-node-port"
  network = google_compute_network.this.id

  allow {
    protocol = "tcp"
    ports    = [tostring(var.node_port)]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = [local.node_tag]
}
