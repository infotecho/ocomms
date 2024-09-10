terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.2"
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
