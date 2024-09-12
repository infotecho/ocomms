resource "google_project_service" "billing" {
  service = "cloudbilling.googleapis.com"
}

resource "google_project_service" "budget" {
  service = "billingbudgets.googleapis.com"
}

data "google_billing_account" "otech" {
  depends_on      = [google_project_service.billing, google_project_service.budget]
  billing_account = "0141DF-FABD06-51DA98"
}

resource "google_billing_budget" "budget" {
  billing_account = data.google_billing_account.otech.id
  display_name    = "O-Comms Budget"

  budget_filter {
    projects = ["projects/${data.google_project.ocomms.number}"]
  }

  amount {
    specified_amount {
      currency_code = "CAD"
      units         = "50"
    }
  }

  threshold_rules {
    threshold_percent = 0.5
    spend_basis       = "CURRENT_SPEND"
  }
  threshold_rules {
    threshold_percent = 0.99
    spend_basis       = "CURRENT_SPEND"
  }
  threshold_rules {
    threshold_percent = 1.0
    spend_basis       = "FORECASTED_SPEND"
  }
}
