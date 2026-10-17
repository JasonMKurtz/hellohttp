variable "project" {
  type    = string
  default = "hellohttp-a"
}

# Madrid.
variable "region" {
  type    = string
  default = "europe-southwest1"
}

variable "cluster_name" {
  type    = string
  default = "hellohttp"
}

# Off switch for everything that costs money. false removes the cluster,
# which also removes its nodes and the load balancer. The network, service
# account, and Artifact Registry repo (free or near-free) are kept.
variable "cluster_enabled" {
  type    = bool
  default = true
}

variable "subnet_cidr" {
  type    = string
  default = "10.0.0.0/20"
}

# Tag of the hellohttp image in Artifact Registry to deploy when the
# Deployment is first created. After that, you own the image (see k8s.tf).
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

# The "Personal" billing account that hellohttp-a is linked to
# (`gcloud billing projects describe hellohttp-a`).
variable "billing_account" {
  type    = string
  default = "01E592-95E34C-CF77B4"
}
