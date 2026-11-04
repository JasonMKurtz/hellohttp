terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.13"
    }
    # For the Karpenter NodePool (nodepool.tf). Unlike the kubernetes
    # provider's kubernetes_manifest, it doesn't need the cluster to exist at
    # plan time, so the cluster can still be created in the same apply.
    kubectl = {
      source  = "alekc/kubectl"
      version = "~> 2.4"
    }
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project   = "hellohttp"
      ManagedBy = "terraform"
    }
  }
}
