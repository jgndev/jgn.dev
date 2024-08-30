# main.tf

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  region      = var.region
  credentials = base64decode(var.credentials)
  project     = var.project_id
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
  project            = var.project_id
  service            = each.key
  disable_on_destroy = false
}

# Create Cloud Storage bucket for posts
resource "google_storage_bucket" "posts_bucket" {
  name                        = "${var.project_id}-posts"
  location                    = var.region
  project                     = var.project_id
  uniform_bucket_level_access = true
  depends_on                  = [google_project_service.services]
}

# Create Pub/Sub topic
resource "google_pubsub_topic" "bucket_changes" {
  name       = "bucket-changes"
  project    = var.project_id
  depends_on = [google_project_service.services]
}

# Create Firestore database
resource "google_firestore_database" "database" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"
  depends_on  = [google_project_service.services]
}

# Create Cloud Run service for the website
resource "google_cloud_run_service" "website" {
  name     = "jgn-website"
  location = var.region
  project  = var.project_id

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
  project  = var.project_id

  metadata {
    namespace = var.project_id
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
  project  = var.project_id
  role     = "roles/run.invoker"
  member   = "allUsers"

  depends_on = [google_cloud_run_service.website]
}

