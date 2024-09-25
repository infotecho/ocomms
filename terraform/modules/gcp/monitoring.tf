resource "google_monitoring_notification_channel" "email_sre" {
  display_name = "Email SRE"
  type         = "email"
  labels = {
    email_address = "sre@infotechottawa.ca"
  }
  force_delete = false
}

resource "google_monitoring_alert_policy" "error_logs" {
  display_name = "Error Logs"
  combiner     = "OR"
  conditions {
    display_name = "Log severity is ERROR or worse"
    condition_matched_log {
      filter = "severity >= ERROR"
    }
  }
  notification_channels = [google_monitoring_notification_channel.email_sre.name]
  alert_strategy {
    notification_rate_limit {
      period = "300s" // 5 minutes
    }
  }
}
