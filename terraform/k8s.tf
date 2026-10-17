# Kubernetes objects managed by Terraform. Because they depend on the
# cluster, `terraform destroy` (or cluster_enabled=false) deletes them first,
# and Auto Mode removes the AWS load balancer before the cluster goes away.

# When cluster_enabled flips to false, aws_eks_cluster.this[0] vanishes from
# the config in that same plan, but Terraform still has to connect to the
# cluster to delete the objects below. So fall back to reading the live
# cluster directly whenever it exists.
data "aws_eks_clusters" "all" {}

data "aws_eks_cluster" "live" {
  count = contains(data.aws_eks_clusters.all.names, var.cluster_name) ? 1 : 0

  name = var.cluster_name
}

locals {
  # The resource attribute is "unknown" while the cluster is first being
  # created, which makes Terraform wait to configure the provider until it is.
  cluster_endpoint = try(aws_eks_cluster.this[0].endpoint, data.aws_eks_cluster.live[0].endpoint, "https://cluster-disabled.invalid")
  cluster_ca       = try(aws_eks_cluster.this[0].certificate_authority[0].data, data.aws_eks_cluster.live[0].certificate_authority[0].data, "")
}

provider "kubernetes" {
  host                   = local.cluster_endpoint
  cluster_ca_certificate = base64decode(local.cluster_ca)

  # Fetch a fresh token on every run instead of storing one.
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", var.cluster_name, "--region", var.region]
  }
}

locals {
  k8s_count      = var.cluster_enabled ? 1 : 0
  app_labels     = { app = "hellohttp" }
  container_port = 5000 # src/hello/routes.go
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

    template {
      metadata {
        labels = local.app_labels
      }

      spec {
        container {
          name  = "hellohttp"
          image = "${aws_ecr_repository.hellohttp.repository_url}:${var.image_tag}"

          port {
            container_port = local.container_port
          }

          env {
            name = "MY_IMAGE"
            value_from {
              field_ref {
                field_path = "spec.nodeName"
              }
            }
          }

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

  # CI deploys new versions with `kubectl set image`; don't roll them back.
  # image_tag only sets the image when the Deployment is first created.
  lifecycle {
    ignore_changes = [spec[0].template[0].spec[0].container[0].image]
  }

  depends_on = [time_sleep.access_propagation]
}

resource "kubernetes_service_v1" "hellohttp" {
  count = local.k8s_count

  metadata {
    name = "hellohttp"
  }

  spec {
    type     = "ClusterIP"
    selector = local.app_labels

    port {
      port        = 80
      target_port = local.container_port
    }
  }

  depends_on = [time_sleep.access_propagation]
}

# Hands Ingresses to EKS Auto Mode's built-in ALB controller.
resource "kubernetes_ingress_class_v1" "alb" {
  count = local.k8s_count

  metadata {
    name = "alb"
    annotations = {
      "ingressclass.kubernetes.io/is-default-class" = "true"
    }
  }

  spec {
    controller = "eks.amazonaws.com/alb"
  }

  depends_on = [time_sleep.access_propagation]
}

# Creates an internet-facing Application Load Balancer.
resource "kubernetes_ingress_v1" "hellohttp" {
  count = local.k8s_count

  metadata {
    name = "hellohttp"
    annotations = {
      "alb.ingress.kubernetes.io/scheme"           = "internet-facing"
      "alb.ingress.kubernetes.io/healthcheck-path" = "/version"
    }
  }

  spec {
    ingress_class_name = kubernetes_ingress_class_v1.alb[0].metadata[0].name

    default_backend {
      service {
        name = kubernetes_service_v1.hellohttp[0].metadata[0].name
        port {
          number = 80
        }
      }
    }
  }

  # Wait for the ALB so its hostname is available as an output.
  wait_for_load_balancer = true
}
