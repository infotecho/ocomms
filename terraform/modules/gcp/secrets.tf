resource "google_secret_manager_secret" "twilio_api_key" {
  secret_id = "twilio-api-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_iam_member" "default" {
  secret_id = google_secret_manager_secret.twilio_api_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.ocomms.email}"
}
