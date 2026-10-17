# Lets GitHub Actions in JasonMKurtz/hellohttp assume an AWS role using
# short-lived OIDC tokens, so no AWS keys are stored in GitHub.

# One per AWS account. AWS validates GitHub's certificate itself, so the
# thumbprint isn't needed.
resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
}

resource "aws_iam_role" "github_deploy" {
  name = "${var.cluster_name}-github-deploy"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          # Only workflows running on master, not PRs or other branches.
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_repo}:ref:refs/heads/master"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "github_deploy" {
  name = "push-and-deploy"
  role = aws_iam_role.github_deploy.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "ecr:GetAuthorizationToken"
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:CompleteLayerUpload",
          "ecr:GetDownloadUrlForLayer",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage",
          "ecr:UploadLayerPart",
        ]
        Resource = aws_ecr_repository.hellohttp.arn
      },
      {
        # For `aws eks update-kubeconfig` / `get-token`.
        Effect   = "Allow"
        Action   = "eks:DescribeCluster"
        Resource = "arn:aws:eks:${var.region}:*:cluster/${var.cluster_name}"
      },
    ]
  })
}

# Kubernetes access for CI: edit rights in the default namespace only,
# enough for `kubectl set image` and `kubectl rollout status`.
resource "aws_eks_access_entry" "github_deploy" {
  count = var.cluster_enabled ? 1 : 0

  cluster_name  = aws_eks_cluster.this[0].name
  principal_arn = aws_iam_role.github_deploy.arn
}

resource "aws_eks_access_policy_association" "github_deploy" {
  count = var.cluster_enabled ? 1 : 0

  cluster_name  = aws_eks_cluster.this[0].name
  principal_arn = aws_eks_access_entry.github_deploy[0].principal_arn
  policy_arn    = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy"

  access_scope {
    type       = "namespace"
    namespaces = ["default"]
  }
}
