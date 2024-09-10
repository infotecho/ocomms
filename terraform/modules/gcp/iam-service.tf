resource "google_service_account" "ocomms" {
  depends_on = [google_project_service.iam]
  account_id = "ocomms"
}

// Allows CD pipeline to associate Cloud Run service to its service account during deployment
resource "google_service_account_iam_binding" "ci_service_account_user" {
  service_account_id = google_service_account.ocomms.name
  role               = "roles/iam.serviceAccountUser"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${var.github_repo_name}"
  ]
}
