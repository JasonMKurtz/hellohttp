resource "google_artifact_registry_repository" "hellohttp" {
  repository_id = "hellohttp"
  location      = var.region
  format        = "DOCKER"

  # Keep the registry from growing forever: the 20 most recent images stay,
  # older ones are deleted.
  cleanup_policy_dry_run = false

  cleanup_policies {
    id     = "keep-recent"
    action = "KEEP"
    most_recent_versions {
      keep_count = 20
    }
  }

  cleanup_policies {
    id     = "delete-rest"
    action = "DELETE"
    condition {
      tag_state = "ANY"
    }
  }

  depends_on = [google_project_service.this]
}
