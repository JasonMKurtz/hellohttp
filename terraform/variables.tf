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

# The load balancer's front end. dualstack-without-public-ipv4 is IPv6 only
# (and saves its public IPv4 charges); dualstack adds IPv4 so clients
# without IPv6 can connect too. Changing it updates the ALB in place.
variable "alb_ip_address_type" {
  type    = string
  default = "dualstack-without-public-ipv4"

  validation {
    condition     = contains(["dualstack-without-public-ipv4", "dualstack"], var.alb_ip_address_type)
    error_message = "alb_ip_address_type must be dualstack-without-public-ipv4 (IPv6 only) or dualstack (IPv4 + IPv6)."
  }
}

# Tag of the hellohttp image in ECR to deploy when the Deployment is first
# created. After that, CI owns the image (see k8s.tf).
variable "image_tag" {
  type    = string
  default = "latest"
}

variable "replicas" {
  type    = number
  default = 2
}

variable "budget_limit_usd" {
  type    = string
  default = "10"
}

# Where budget alerts go. Set it in terraform.tfvars (gitignored) to keep
# the address out of the repo.
variable "budget_email" {
  type = string
}

# GitHub repo whose master-branch workflows may push images and deploy.
variable "github_repo" {
  type    = string
  default = "JasonMKurtz/hellohttp"
}

# IAM users/roles that get cluster-admin via EKS access entries.
variable "admin_principal_arns" {
  type    = list(string)
  default = ["arn:aws:iam::162612267003:user/JasonKurtz"]
}
