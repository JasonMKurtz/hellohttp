variable "region" {
  type    = string
  default = "eu-north-1"
}

variable "cluster_name" {
  type    = string
  default = "hellohttp"
}

# Off switch for everything that costs money. false removes the cluster,
# which also terminates its nodes and releases their public IPs. The VPC,
# IAM roles, and ECR repo (free or near-free) are kept.
variable "cluster_enabled" {
  type    = bool
  default = true
}

variable "kubernetes_version" {
  type    = string
  default = "1.36"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

# Tag of the hellohttp image in ECR to deploy.
variable "image_tag" {
  type    = string
  default = "latest"
}

variable "replicas" {
  type    = number
  default = 2
}

# IAM users/roles that get cluster-admin via EKS access entries.
variable "admin_principal_arns" {
  type    = list(string)
  default = ["arn:aws:iam::162612267003:user/JasonKurtz"]
}
