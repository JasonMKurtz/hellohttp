output "cluster_name" {
  value = one(aws_eks_cluster.this[*].name)
}

output "ecr_repository_url" {
  value = aws_ecr_repository.hellohttp.repository_url
}

output "app_url" {
  value = var.cluster_enabled ? "http://${try(kubernetes_ingress_v1.hellohttp[0].status[0].load_balancer[0].ingress[0].hostname, "")}" : null
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "aws eks update-kubeconfig --region ${var.region} --name ${var.cluster_name}" : null
}
