resource "google_project_service" "artifact_registry" {
  service = "artifactregistry.googleapis.com"
}

resource "google_artifact_registry_repository" "ocomms" {
  depends_on    = [google_project_service.artifact_registry]
  repository_id = "ocomms"
  location      = "northamerica-northeast1"
  format        = "DOCKER"
}

resource "google_artifact_registry_repository_iam_binding" "github" {
  repository = google_artifact_registry_repository.ocomms.repository_id
  location   = google_artifact_registry_repository.ocomms.location
  role       = "roles/artifactregistry.writer"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${var.github_repo_name}"
  ]
}
