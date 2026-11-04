output "cluster_name" {
  value = one(aws_eks_cluster.this[*].name)
}

output "ecr_repository_url" {
  value = aws_ecr_repository.hellohttp.repository_url
}

output "github_deploy_role_arn" {
  value = aws_iam_role.github_deploy.arn
}

# The /56 AWS gave the VPC.
output "vpc_ipv6_cidr" {
  value = aws_vpc.this.ipv6_cidr_block
}

# The load balancer's DNS name. With the default (IPv6-only) front end it
# only resolves to IPv6 addresses: test with `curl -6` from a network that
# has IPv6.
output "app_url" {
  value = var.cluster_enabled ? "http://${try(kubernetes_ingress_v1.hellohttp[0].status[0].load_balancer[0].ingress[0].hostname, "")}" : null
}

output "configure_kubectl" {
  value = var.cluster_enabled ? "aws eks update-kubeconfig --region ${var.region} --name ${var.cluster_name}" : null
}
