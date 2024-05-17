terraform {
  cloud {
    organization = "infotecho"
    workspaces {
      name = "ocomms"
    }
  }
}

module "gcp" {
  source = "./gcp"
}
