# Kubernetes objects managed by Terraform. Because they depend on the
# cluster, `terraform destroy` (or cluster_enabled=false) deletes them first,
# and GKE removes any load balancer before the cluster goes away.

# Your gcloud login (application default credentials), which GKE also
# accepts as a Kubernetes token. No CLI plugin needed, unlike EKS.
data "google_client_config" "default" {}

# When cluster_enabled flips to false, google_container_cluster.this[0]
# vanishes from the config in that same plan, but Terraform still has to
# connect to the cluster to delete the objects below. So fall back to
# reading the live cluster from the GKE API whenever it exists. (The
# google_container_cluster data source errors if the cluster is missing.)
data "http" "live_cluster" {
  url = "https://container.googleapis.com/v1/projects/${var.project}/locations/${var.zone}/clusters/${var.cluster_name}"

  request_headers = {
    Authorization       = "Bearer ${data.google_client_config.default.access_token}"
    x-goog-user-project = var.project
  }
}

locals {
  live_cluster = data.http.live_cluster.status_code == 200 ? jsondecode(data.http.live_cluster.response_body) : null

  # The resource attribute is "unknown" while the cluster is first being
  # created, which makes Terraform wait to configure the provider until it is.
  cluster_endpoint = try("https://${google_container_cluster.this[0].endpoint}", "https://${local.live_cluster.endpoint}", "https://cluster-disabled.invalid")
  cluster_ca       = try(google_container_cluster.this[0].master_auth[0].cluster_ca_certificate, local.live_cluster.masterAuth.clusterCaCertificate, "")
}

provider "kubernetes" {
  host                   = local.cluster_endpoint
  cluster_ca_certificate = base64decode(local.cluster_ca)
  token                  = data.google_client_config.default.access_token
}

locals {
  k8s_count      = var.cluster_enabled ? 1 : 0
  ingress_count  = var.cluster_enabled && var.ingress_enabled ? 1 : 0
  app_labels     = { app = "hellohttp" }
  container_port = 5000 # passed to the app as PORT (its default is 8080)
  image_repo     = "${var.region}-docker.pkg.dev/${var.project}/${google_artifact_registry_repository.hellohttp.repository_id}/hellohttp"
}

resource "kubernetes_deployment_v1" "hellohttp" {
  count = local.k8s_count

  metadata {
    name   = "hellohttp"
    labels = local.app_labels
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = local.app_labels
    }

    # Zero-downtime rollouts, part 1: replace one pod at a time and never
    # drop below `replicas` ready pods, regardless of how percentages would
    # round.
    strategy {
      type = "RollingUpdate"
      rolling_update {
        max_surge       = 1
        max_unavailable = 0
      }
    }

    template {
      metadata {
        labels = local.app_labels
      }

      spec {
        # Part 3 lives in the app: on SIGTERM it keeps serving for 15s while
        # kube-proxy (or the load balancer) stops sending it traffic, then
        # drains for up to 10s (Listen() in src/lib/routes). The grace period
        # must cover both, or Kubernetes kills the process mid-drain.
        termination_grace_period_seconds = 30

        container {
          name  = "hellohttp"
          image = "${local.image_repo}:${var.image_tag}"

          port {
            container_port = local.container_port
          }

          env {
            name  = "PORT"
            value = tostring(local.container_port)
          }

          env {
            name = "MY_IMAGE"
            value_from {
              field_ref {
                field_path = "spec.nodeName"
              }
            }
          }

          # What the scheduler reserves on the node for each pod. Kept small
          # so two pods, plus a third during a rollout, fit on one small node.
          resources {
            requests = {
              cpu    = "10m"
              memory = "32Mi"
            }
            limits = {
              memory = "64Mi"
            }
          }

          # Only Ready pods get traffic. A load balancer's health check also
          # copies this path.
          readiness_probe {
            http_get {
              path = "/version"
              port = local.container_port
            }
          }
        }
      }
    }
  }

  # Don't block the apply waiting for pods (e.g. before an image is pushed).
  wait_for_rollout = false

  # New versions are deployed with `kubectl set image`; don't roll them
  # back. image_tag only sets the image when the Deployment is first created.
  lifecycle {
    ignore_changes = [spec[0].template[0].spec[0].container[0].image]
  }
}

# A NodePort Service: every node listens on node_port and forwards to a
# Ready pod, on whichever node it runs. It also gets a ClusterIP for traffic
# inside the cluster, and an Ingress can use it as a backend.
resource "kubernetes_service_v1" "hellohttp" {
  count = local.k8s_count

  metadata {
    name = "hellohttp"
  }

  spec {
    type     = "NodePort"
    selector = local.app_labels

    port {
      port        = 80
      target_port = local.container_port
      node_port   = var.node_port
    }
  }

  # GKE adds `cloud.google.com/neg: {"ingress":true}` to every Service on
  # VPC-native clusters: container-native load balancing, where a load
  # balancer sends traffic to pod IPs directly (through a "network endpoint
  # group") instead of bouncing through nodes. Part 2: with an Ingress, GKE
  # then adds a readiness gate to each pod, so a new pod only counts as
  # Ready once the load balancer's health check passes, not just its own
  # readiness probe. (On EKS this took a namespace label.)
  lifecycle {
    ignore_changes = [metadata[0].annotations["cloud.google.com/neg"]]
  }
}

# GKE's built-in Ingress controller turns this into an external Application
# Load Balancer with a public IP. It takes a few minutes after the IP shows
# up before requests stop returning 404/502.
resource "kubernetes_ingress_v1" "hellohttp" {
  count = local.ingress_count

  metadata {
    name = "hellohttp"
    annotations = {
      "kubernetes.io/ingress.class" = "gce"
    }
  }

  spec {
    default_backend {
      service {
        name = kubernetes_service_v1.hellohttp[0].metadata[0].name
        port {
          number = 80
        }
      }
    }
  }

  # Wait for the load balancer so its IP is available as an output.
  wait_for_load_balancer = true
}

# For the app_url output: the node's public IP. It changes when GKE
# replaces the node (upgrades, repairs), so re-run `terraform refresh`.
data "kubernetes_nodes" "all" {
  count = local.k8s_count

  depends_on = [google_container_node_pool.default]
}
