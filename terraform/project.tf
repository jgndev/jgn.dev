# provider "google" {
#   region      = var.region
#   credentials = base64decode(var.credentials)
# }

resource "google_project" "website_project" {
  name            = "jgn website"
  project_id      = var.project_id
  org_id          = var.org_id
  billing_account = var.billing_account
}

resource "google_project_service" "cloudbilling" {
  project = google_project.website_project.project_id
  service = "cloudbilling.googleapis.com"

  disable_on_destroy = false
  depends_on         = [google_project.website_project]
}

output "project_id" {
  value = google_project.website_project.project_id
}
