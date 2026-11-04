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

# The node's public IPv6 address and the NodePort. IPv6 addresses go in
# brackets in a URL. Test with `curl -6`; check first that your own network
# has IPv6 (`curl -6 https://ifconfig.co`).
output "app_url_ipv6" {
  value = !var.cluster_enabled ? null : (
    "http://[${try(data.google_compute_instance.node[0].network_interface[0].ipv6_access_config[0].external_ipv6, "")}]:${var.node_port}"
  )
}

# The /64 Google gave the subnet.
output "subnet_ipv6_prefix" {
  value = google_compute_subnetwork.this.external_ipv6_prefix
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "gcloud container clusters get-credentials ${var.cluster_name} --zone ${var.zone} --project ${var.project}" : null
}
