variable "project" {
  type    = string
  default = "hellohttp-a"
}

# Among the cheapest regions, and where the free tier's 30 GB disk applies
# (also us-west1 and us-east1).
variable "region" {
  type    = string
  default = "us-central1"
}

# A zonal cluster (one zone) instead of a regional one: GKE's free tier
# covers the management fee of one zonal or Autopilot cluster, and the
# nodes don't get tripled across three zones.
variable "zone" {
  type    = string
  default = "us-central1-a"
}

# The smallest machine GKE runs on. e2-micro (free tier) was tried: its
# ~600 MiB for pods isn't enough even for GKE's own system pods, so
# kube-dns never scheduled.
variable "node_machine_type" {
  type    = string
  default = "e2-small"
}

# Spot VMs are 60-90% cheaper (~$4/month for an e2-small) but Google can
# stop them at any time; GKE then brings up a replacement node.
variable "node_spot" {
  type    = bool
  default = true
}

# The load balancer has no free tier (~$18/month). Without it, the app is
# reached on node_port of the node's public IP. Turn it on to try the
# Ingress; you pay by the hour, so turn it back off afterwards.
variable "ingress_enabled" {
  type    = bool
  default = false
}

# Port the app is reachable on, on every node (30000-32767).
variable "node_port" {
  type    = number
  default = 30080
}

variable "cluster_name" {
  type    = string
  default = "hellohttp"
}

# Off switch for the cluster. false removes it, which also removes its node
# and the load balancer (if ingress_enabled). The network, service account,
# and Artifact Registry repo (free or near-free) are kept.
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
