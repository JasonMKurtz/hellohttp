# GCP services are off until enabled per project. Left on when Terraform
# destroys this stack: other things in the project may use them, and
# disabling one can delete resources Terraform doesn't know about.
resource "google_project_service" "this" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "billingbudgets.googleapis.com",
    "cloudbuild.googleapis.com", # builds the image (see README)
    "compute.googleapis.com",
    "container.googleapis.com",
  ])

  service            = each.key
  disable_on_destroy = false
}
