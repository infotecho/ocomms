resource "google_project_service" "compute" {
  service = "compute.googleapis.com"
}

resource "google_compute_region_network_endpoint_group" "ocomms" {
  depends_on            = [google_project_service.compute]
  name                  = "ocomms"
  network_endpoint_type = "SERVERLESS"
  region                = "northamerica-northeast1"
  cloud_run {
    service = "ocomms"
  }
}

module "lb-http" {
  source  = "terraform-google-modules/lb-http/google//modules/serverless_negs"
  version = "~> 11.0"

  name    = "ocomms"
  project = data.google_project.ocomms.name

  ssl                             = true
  managed_ssl_certificate_domains = ["ocomms.infotechottawa.ca"]
  https_redirect                  = true

  backends = {
    default = {
      groups = [
        {
          group = google_compute_region_network_endpoint_group.ocomms.id
        }
      ]
      enable_cdn = false
      iap_config = {
        enable = false
      }
      log_config = {
        enable = false
      }
    }
  }
}
