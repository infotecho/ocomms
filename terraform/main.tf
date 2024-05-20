terraform {
  cloud {
    organization = "infotecho"
    workspaces {
      name = "ocomms"
    }
  }
}

module "github" {
  source = "./modules/github"
}

module "gcp" {
  source           = "./modules/gcp"
  github_repo_name = module.github.repo_full_name
}
