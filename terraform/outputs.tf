output "cluster_name" {
  value = one(aws_eks_cluster.this[*].name)
}

output "ecr_repository_url" {
  value = aws_ecr_repository.hellohttp.repository_url
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "aws eks update-kubeconfig --region ${var.region} --name ${var.cluster_name}" : null
}
