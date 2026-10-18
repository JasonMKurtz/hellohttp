# Node service account: what the nodes run as. The default (the
# Compute Engine default service account) has Owner on this project, far
# more than nodes need.
resource "google_service_account" "node" {
  account_id   = "${var.cluster_name}-node"
  display_name = "${var.cluster_name} GKE nodes"
}

# Logging, monitoring, and the other basics every node needs.
resource "google_project_iam_member" "node" {
  project = var.project
  role    = "roles/container.defaultNodeServiceAccount"
  member  = google_service_account.node.member
}

# Pull images from our repo only.
resource "google_artifact_registry_repository_iam_member" "node" {
  repository = google_artifact_registry_repository.hellohttp.name
  location   = google_artifact_registry_repository.hellohttp.location
  role       = "roles/artifactregistry.reader"
  member     = google_service_account.node.member
}

# There's no equivalent of EKS access entries: GKE maps Google identities to
# Kubernetes users, and project IAM roles grant cluster access. Your Owner
# role makes you a cluster admin.
