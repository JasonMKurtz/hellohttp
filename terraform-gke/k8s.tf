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
  k8s_count     = var.cluster_enabled ? 1 : 0
  ingress_count = var.cluster_enabled && var.ingress_enabled ? 1 : 0
  registry      = "${var.region}-docker.pkg.dev/${var.project}/${google_artifact_registry_repository.hellohttp.repository_id}"
  image_repo    = "${local.registry}/hellohttp"
  app_port      = 5000 # passed to every app as PORT (their default is 8080)

  # The apps to run: one Deployment and one Service each, named after the
  # key and labeled app = <key>. Only what differs between them is here.
  apps = {
    hellohttp = {
      image      = "hellohttp"
      probe_path = "/version"
      replicas   = var.replicas
      node_port  = var.node_port # also reachable on the node's public IP
    }
    hellohttp-backend = {
      image      = "hellohttp-backend"
      probe_path = "/api/foo" # its only route, so also the LB health check
      replicas   = 1
      node_port  = null # ClusterIP only
    }
  }

  # No apps when the cluster is off.
  enabled_apps = var.cluster_enabled ? local.apps : {}
}

resource "kubernetes_deployment_v1" "app" {
  for_each = local.enabled_apps

  metadata {
    name   = each.key
    labels = { app = each.key }
  }

  spec {
    replicas = each.value.replicas

    selector {
      match_labels = { app = each.key }
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
        labels = { app = each.key }
      }

      spec {
        # Part 3 lives in the app: on SIGTERM it keeps serving for 15s while
        # kube-proxy (or the load balancer) stops sending it traffic, then
        # drains for up to 10s (Listen() in src/lib/routes). The grace period
        # must cover both, or Kubernetes kills the process mid-drain.
        termination_grace_period_seconds = 30

        container {
          name  = each.key
          image = "${local.registry}/${each.value.image}:${var.image_tag}"

          port {
            container_port = local.app_port
          }

          env {
            name  = "PORT"
            value = tostring(local.app_port)
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
          # so every app's pods, plus one extra during a rollout, fit on one
          # small node.
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
          # copies this path; without a probe it checks GET /.
          readiness_probe {
            http_get {
              path = each.value.probe_path
              port = local.app_port
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

# With a node_port, a NodePort Service: every node listens on it and
# forwards to a Ready pod, on whichever node it runs. Without one, a
# ClusterIP Service, reachable only inside the cluster and by the load
# balancer. Either way, an Ingress can use it as a backend.
resource "kubernetes_service_v1" "app" {
  for_each = local.enabled_apps

  metadata {
    name = each.key
  }

  spec {
    type     = each.value.node_port == null ? "ClusterIP" : "NodePort"
    selector = { app = each.key }

    port {
      port        = 80
      target_port = local.app_port
      node_port   = each.value.node_port
    }
  }

  # GKE adds `cloud.google.com/neg: {"ingress":true}` to every Service on
  # VPC-native clusters: container-native load balancing, where a load
  # balancer sends traffic to pod IPs directly (through a "network endpoint
  # group") instead of bouncing through nodes. Part 2: with an Ingress, GKE
  # then adds a readiness gate to each pod, so a new pod only counts as
  # Ready once the load balancer's health check passes, not just its own
  # readiness probe. (On EKS this took a namespace label.) With an Ingress,
  # GKE also records which endpoint groups it made in neg-status.
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["cloud.google.com/neg"],
      metadata[0].annotations["cloud.google.com/neg-status"],
    ]
  }
}

# These were single resources before there was more than one app. Moving
# them keeps Terraform from deleting and recreating the running ones.
moved {
  from = kubernetes_deployment_v1.hellohttp[0]
  to   = kubernetes_deployment_v1.app["hellohttp"]
}

moved {
  from = kubernetes_service_v1.hellohttp[0]
  to   = kubernetes_service_v1.app["hellohttp"]
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
        name = kubernetes_service_v1.app["hellohttp"].metadata[0].name
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
