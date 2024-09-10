terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.2"
    }
  }
}

provider "github" {
  owner = "infotecho"
}

resource "github_repository" "ocomms" {
  name       = "ocomms"
  visibility = "private"
  has_issues = true
}
