resource "google_project_service" "iam" {
  service = "iam.googleapis.com"
}

resource "google_service_account" "ocomms" {
  depends_on = [google_project_service.iam]
  account_id = "ocomms"
}

resource "google_secret_manager_secret_iam_member" "ocomms-twilio" {
  secret_id = google_secret_manager_secret.twilio_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.ocomms.email}"
}

resource "google_secret_manager_secret_iam_member" "ocomms-sendgrid" {
  secret_id = google_secret_manager_secret.sendgrid_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.ocomms.email}"
}
