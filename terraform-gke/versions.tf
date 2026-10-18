terraform {
  required_version = ">= 1.10"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 3.0"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

# Credentials come from `gcloud auth application-default login`.
provider "google" {
  project = var.project
  region  = var.region

  # Bill API quota to this project. The Budgets API refuses user credentials
  # without one.
  user_project_override = true
  billing_project       = var.project

  default_labels = {
    project    = "hellohttp"
    managed-by = "terraform"
  }
}
