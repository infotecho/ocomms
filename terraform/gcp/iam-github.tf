resource "google_iam_workload_identity_pool" "ci" {
  workload_identity_pool_id = "ci-cd"
  display_name              = "CI/CD Runners"
  description               = "Workload identity pool for CI/CD runners"
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.ci.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub"
  description                        = "Workload identity pool provider for GitHub-hosted runners"
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }
  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account" "github" {
  account_id   = "github"
  display_name = "GitHub Actions Runner"
}

resource "google_service_account_iam_binding" "name" {
  service_account_id = google_service_account.github.name
  role               = "roles/iam.workloadIdentityUser"
  members = [
    "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/infotecho/ocomms"
  ]
}
