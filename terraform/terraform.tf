# terraform.tf

terraform {
  cloud {
    organization = "jgndev"
    workspaces {
      name = "jgndev"
    }
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

