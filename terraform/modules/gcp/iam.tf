resource "google_project_service" "iam" {
  service = "iam.googleapis.com"
}

resource "google_service_account" "ocomms" {
  depends_on = [google_project_service.iam]
  account_id = "ocomms"
}
