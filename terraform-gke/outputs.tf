output "cluster_name" {
  value = one(google_container_cluster.this[*].name)
}

output "image_repo" {
  value = local.image_repo
}

locals {
  node_ips = flatten([
    for node in try(data.kubernetes_nodes.all[0].nodes, []) : [
      for addr in node.status[0].addresses : addr.address if addr.type == "ExternalIP"
    ]
  ])
}

# The load balancer's IP with ingress_enabled, otherwise the first node's
# public IP and the NodePort.
output "app_url" {
  value = !var.cluster_enabled ? null : (
    var.ingress_enabled
    ? "http://${try(kubernetes_ingress_v1.hellohttp[0].status[0].load_balancer[0].ingress[0].ip, "")}"
    : "http://${try(local.node_ips[0], "")}:${var.node_port}"
  )
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "gcloud container clusters get-credentials ${var.cluster_name} --zone ${var.zone} --project ${var.project}" : null
}
