# variables.tf

variable "project_id" {
  description = "The ID of the Google Cloud project"
  type        = string
}

variable "region" {
  description = "The region to deploy resources to"
  type        = string
  default     = "us-central1"
}

variable "org_id" {
  description = "The organization ID"
  type        = string
}

variable "billing_account" {
  description = "The billing account ID"
  type        = string
}