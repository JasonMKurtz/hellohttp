variable "region" {
  type    = string
  default = "eu-north-1"
}

variable "cluster_name" {
  type    = string
  default = "hellohttp"
}

variable "kubernetes_version" {
  type    = string
  default = "1.36"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

# IAM users/roles that get cluster-admin via EKS access entries.
variable "admin_principal_arns" {
  type    = list(string)
  default = ["arn:aws:iam::162612267003:user/JasonKurtz"]
}
