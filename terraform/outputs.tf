output "cluster_name" {
  value = aws_eks_cluster.this.name
}

output "ecr_repository_url" {
  value = aws_ecr_repository.hellohttp.repository_url
}

output "configure_kubectl" {
  value = "aws eks update-kubeconfig --region ${var.region} --name ${aws_eks_cluster.this.name}"
}
