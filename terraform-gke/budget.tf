data "google_project" "this" {}

# Covers the whole billing account, not just this project, so it also
# catches anything created by hand. Kept even when cluster_enabled=false.
# Alerts email the billing account's admins (you). Billing data lags by
# hours, so treat this as a backstop, not a real-time alarm.
resource "google_billing_budget" "monthly" {
  billing_account = data.google_project.this.billing_account
  display_name    = "monthly-cost"

  amount {
    specified_amount {
      currency_code = "USD"
      units         = var.budget_limit_usd
    }
  }

  threshold_rules {
    threshold_percent = 0.8
  }

  # Warns early in the month if spend is on track to exceed the limit.
  threshold_rules {
    threshold_percent = 1.0
    spend_basis       = "FORECASTED_SPEND"
  }

  depends_on = [google_project_service.this]
}
