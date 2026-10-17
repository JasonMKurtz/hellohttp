# GCP services are off until enabled per project. Left on when Terraform
# destroys this stack: other things in the project may use them, and
# disabling one can delete resources Terraform doesn't know about.
resource "google_project_service" "this" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "billingbudgets.googleapis.com",
    "cloudbuild.googleapis.com", # builds the image (see README)
    # budget.tf reads the project during plan, before this resource could
    # turn it on, so the first apply needs it enabled by hand (see README).
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
  ])

  service            = each.key
  disable_on_destroy = false
}
