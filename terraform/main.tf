# main.tf

# Configure the Google Cloud provider
provider "google" {
  project = var.project_id
  region  = var.region
}

# Create a new GCP Project
resource "google_project" "website_project" {
  name            = "JGN Website"
  project_id      = var.project_id
  org_id          = var.org_id
  billing_account = var.billing_account
}

# Enable necessary APIs
resource "google_project_service" "services" {
  for_each = toset([
    "storage.googleapis.com",
    "pubsub.googleapis.com",
    "firestore.googleapis.com",
    "run.googleapis.com",
    "cloudbuild.googleapis.com"
  ])
  project = google_project.website_project.project_id
  service = each.key
  disable_on_destroy = false
}

# Create Cloud Storage bucket for posts
resource "google_storage_bucket" "posts_bucket" {
  name     = "${var.project_id}-posts"
  location = var.region
  project  = google_project.website_project.project_id
  uniform_bucket_level_access = true
}

# Create Pub/Sub topic
resource "google_pubsub_topic" "bucket_changes" {
  name    = "bucket-changes"
  project = google_project.website_project.project_id
}

# Create Firestore database
resource "google_firestore_database" "database" {
  project     = google_project.website_project.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"
}

# Create Cloud Run service for the website
resource "google_cloud_run_service" "website" {
  name     = "jgn-website"
  location = var.region
  project  = google_project.website_project.project_id
  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/jgn-website:latest"
        env {
          name  = "GOOGLE_CLOUD_PROJECT"
          value = var.project_id
        }
      }
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
  depends_on = [google_project_service.services]
}

# Set up custom domain mapping
resource "google_cloud_run_domain_mapping" "domain_mapping" {
  location = var.region
  name     = "jgn.dev"
  project  = google_project.website_project.project_id
  metadata {
    namespace = google_project.website_project.project_id
  }
  spec {
    route_name = google_cloud_run_service.website.name
  }
  depends_on = [google_cloud_run_service.website]
}

# IAM entry for all users to invoke the function
resource "google_cloud_run_service_iam_member" "all_users" {
  service  = google_cloud_run_service.website.name
  location = google_cloud_run_service.website.location
  role     = "roles/run.invoker"
  member   = "allUsers"
  project  = google_project.website_project.project_id
}
