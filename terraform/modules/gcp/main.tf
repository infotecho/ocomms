terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project               = "ocomms"
  region                = "northamerica-northeast1"
  user_project_override = true
  billing_project       = "ocomms"
}

data "google_project" "ocomms" {}
