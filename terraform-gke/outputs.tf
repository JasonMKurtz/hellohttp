output "cluster_name" {
  value = one(google_container_cluster.this[*].name)
}

output "image_repo" {
  value = local.image_repo
}

output "app_url" {
  value = var.cluster_enabled ? "http://${try(kubernetes_ingress_v1.hellohttp[0].status[0].load_balancer[0].ingress[0].ip, "")}" : null
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "gcloud container clusters get-credentials ${var.cluster_name} --region ${var.region} --project ${var.project}" : null
}
