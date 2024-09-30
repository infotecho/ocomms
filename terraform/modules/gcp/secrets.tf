resource "google_secret_manager_secret" "twilio_api_key" {
  secret_id = "twilio-api-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "sendgrid_api_key" {
  secret_id = "sendgrid-api-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "twilio_auth_token" {
  secret_id = "twilio-auth-token"
  replication {
    auto {}
  }
}
