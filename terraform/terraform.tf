# terraform.tf

terraform {
  cloud {
    organization = "jgndev"
    workspaces {
      name = "jgn-dev-08-24-24"
    }
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}