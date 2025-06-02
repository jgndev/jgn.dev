variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP region"
  type        = string
  default     = "us-south1"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "jgn-dev"
}

variable "github_token" {
  description = "GitHub personal access token for API access"
  type        = string
  sensitive   = true
}

variable "github_webhook_secret" {
  description = "Secret for GitHub webhook signature verification"
  type        = string
  sensitive   = true
}

variable "github_owner" {
  description = "GitHub repository owner"
  type        = string
  default     = "jgndev"
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "jgn.dev"
}

variable "custom_domain" {
  description = "Custom domain for the application (optional)"
  type        = string
  default     = "jgn.dev"
}

variable "min_instances" {
  description = "Minimum number of Cloud Run instances"
  type        = number
  default     = 0
}

variable "max_instances" {
  description = "Maximum number of Cloud Run instances"
  type        = number
  default     = 3
}

variable "cpu_limit" {
  description = "CPU limit for Cloud Run container"
  type        = string
  default     = "1"
}

variable "memory_limit" {
  description = "Memory limit for Cloud Run container"
  type        = string
  default     = "512Mi"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "enable_vpc_connector" {
  description = "Enable VPC connector for Cloud Run (for private resources)"
  type        = bool
  default     = false
}

variable "vpc_connector_name" {
  description = "VPC connector name (if enable_vpc_connector is true)"
  type        = string
  default     = ""
}

variable "labels" {
  description = "Labels to apply to resources"
  type        = map(string)
  default = {
    app         = "jgn-dev"
    environment = "production"
    managed-by  = "terraform"
  }
} 