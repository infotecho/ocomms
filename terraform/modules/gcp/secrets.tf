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

resource "google_secret_manager_secret" "primary_agent_did" {
  secret_id = "primary-agent-did"
  replication {
    auto {}
  }
}
