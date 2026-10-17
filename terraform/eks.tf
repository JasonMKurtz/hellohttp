# EKS Auto Mode: AWS manages nodes (Karpenter), load balancers, and EBS storage.
resource "aws_eks_cluster" "this" {
  count = var.cluster_enabled ? 1 : 0

  name     = var.cluster_name
  version  = var.kubernetes_version
  role_arn = aws_iam_role.cluster.arn

  # Auto Mode ships its own networking/DNS/proxy; skip the legacy add-ons.
  bootstrap_self_managed_addons = false

  access_config {
    authentication_mode                         = "API"
    bootstrap_cluster_creator_admin_permissions = false
  }

  compute_config {
    enabled       = true
    node_pools    = ["general-purpose", "system"]
    node_role_arn = aws_iam_role.node.arn
  }

  kubernetes_network_config {
    elastic_load_balancing {
      enabled = true
    }
  }

  storage_config {
    block_storage {
      enabled = true
    }
  }

  vpc_config {
    subnet_ids              = aws_subnet.public[*].id
    endpoint_public_access  = true
    endpoint_private_access = true
  }

  depends_on = [
    aws_iam_role_policy_attachment.cluster,
    aws_route_table_association.public,
  ]
}

# Who can use kubectl. Replaces the console's "creator is admin" default.
resource "aws_eks_access_entry" "admin" {
  for_each = var.cluster_enabled ? toset(var.admin_principal_arns) : toset([])

  cluster_name  = aws_eks_cluster.this[0].name
  principal_arn = each.key
}

resource "aws_eks_access_policy_association" "admin" {
  for_each = aws_eks_access_entry.admin

  cluster_name  = aws_eks_cluster.this[0].name
  principal_arn = each.value.principal_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"

  access_scope {
    type = "cluster"
  }
}
