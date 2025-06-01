# Terraform Infrastructure Guide

This guide covers the Terraform infrastructure for deploying jgn.dev to GCP Cloud Run. Terraform provides Infrastructure as Code (IaC) capabilities, making deployments reproducible and version-controlled.

## 🎯 Overview

The Terraform configuration creates a complete, production-ready infrastructure for hosting jgn.dev on GCP Cloud Run with the following components:

- **Cloud Run Service**: Serverless container hosting
- **Artifact Registry**: Private container image repository
- **Service Accounts**: Secure identity and access management
- **Cloud Build Triggers**: Automated CI/CD integration
- **IAM Bindings**: Least-privilege access controls
- **Domain Mapping**: Custom domain configuration (optional)

## 📁 File Structure

```
terraform/
├── main.tf                   # Main infrastructure configuration
├── variables.tf              # Input variable definitions
├── outputs.tf                # Output value definitions
├── terraform.tfvars.example  # Example configuration file
└── terraform.tfvars          # Your actual configuration (create from example)
```

## 🚀 Quick Start

### Step 1: Prerequisites

Ensure you have the following installed:

```bash
# Install Terraform
brew install terraform  # macOS
# or download from: https://terraform.io/downloads

# Install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Authenticate with GCP
gcloud auth login
gcloud auth application-default login
```

### Step 2: Initialize Configuration

```bash
# Navigate to terraform directory
cd terraform

# Copy example configuration
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
nano terraform.tfvars
```

### Step 3: Deploy Infrastructure

```bash
# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the infrastructure
terraform apply
```

## 🔧 Configuration

### Required Variables

Edit `terraform.tfvars` with your configuration:

```hcl
# GCP Configuration
project_id = "your-gcp-project-id"
region     = "us-central1"

# Application Configuration
app_name = "jgn-dev"

# GitHub Configuration
github_token          = "ghp_your_github_personal_access_token"
github_webhook_secret = "your_webhook_secret_here"
github_owner          = "jgndev"
github_repo           = "jgn.dev"

# Optional: Custom Domain
custom_domain = "jgn.dev"

# Environment
environment = "prod"
```

### Variable Reference

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `project_id` | string | - | **Required**: GCP project ID |
| `region` | string | us-central1 | GCP region for resources |
| `app_name` | string | jgn-dev | Application name |
| `github_token` | string | - | **Required**: GitHub token for API access |
| `github_webhook_secret` | string | "" | GitHub webhook secret |
| `github_owner` | string | jgndev | GitHub repository owner |
| `github_repo` | string | jgn.dev | GitHub repository name |
| `custom_domain` | string | "" | Custom domain (optional) |
| `min_instances` | number | 0 | Minimum Cloud Run instances |
| `max_instances` | number | 10 | Maximum Cloud Run instances |
| `cpu_limit` | string | "1" | CPU limit per instance |
| `memory_limit` | string | "512Mi" | Memory limit per instance |
| `environment` | string | prod | Environment name |

## 🏗️ Infrastructure Components

### Cloud Run Service

The main application hosting service:

```hcl
resource "google_cloud_run_v2_service" "app" {
  name     = var.app_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloud_run_sa.email
    
    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.container_registry.repository_id}/${var.app_name}:latest"
      # ... environment variables and resource limits
    }
  }
}
```

**Features:**
- Auto-scaling from 0 to configured maximum
- Health checks and liveness probes
- Environment variable configuration
- Resource limits and CPU/memory allocation

### Artifact Registry

Private container registry for Docker images:

```hcl
resource "google_artifact_registry_repository" "container_registry" {
  location      = var.region
  repository_id = "${var.app_name}-repo"
  description   = "Container registry for ${var.app_name}"
  format        = "DOCKER"
}
```

**Benefits:**
- Private, secure image storage
- Integration with Cloud Build
- Vulnerability scanning
- Access control via IAM

### Service Account

Dedicated service account with minimal permissions:

```hcl
resource "google_service_account" "cloud_run_sa" {
  account_id   = "${var.app_name}-run-sa"
  display_name = "Cloud Run Service Account for ${var.app_name}"
  description  = "Service account used by ${var.app_name} Cloud Run service"
}
```

**Security:**
- Principle of least privilege
- No over-privileged access
- Dedicated to Cloud Run service only

### Cloud Build Trigger

Automated CI/CD integration:

```hcl
resource "google_cloudbuild_trigger" "github_trigger" {
  name        = "${var.app_name}-deploy-trigger"
  description = "Trigger for deploying ${var.app_name} from GitHub"

  github {
    owner = var.github_owner
    name  = var.github_repo
    push {
      branch = "^main$"
    }
  }
  # ... build steps
}
```

**Capabilities:**
- Automatic builds on GitHub push
- Multi-stage Docker builds
- Automatic deployment to Cloud Run
- Build logging and notifications

## 🔧 Customization

### Environment-Specific Configurations

Create different configurations for different environments:

#### Development Environment

```hcl
# terraform/environments/dev.tfvars
project_id    = "jgn-dev-project"
app_name      = "jgn-dev-staging"
environment   = "dev"
min_instances = 0
max_instances = 3
cpu_limit     = "1"
memory_limit  = "512Mi"
```

#### Production Environment

```hcl
# terraform/environments/prod.tfvars
project_id    = "jgn-prod-project"
app_name      = "jgn-dev"
environment   = "prod"
min_instances = 1
max_instances = 20
cpu_limit     = "2"
memory_limit  = "1Gi"
custom_domain = "jgn.dev"
```

Deploy with specific configuration:

```bash
terraform apply -var-file="environments/prod.tfvars"
```

### Additional Resources

Add monitoring, alerting, or other resources:

```hcl
# Add to main.tf
resource "google_monitoring_alert_policy" "high_error_rate" {
  display_name = "High Error Rate - ${var.app_name}"
  # ... alerting configuration
}

resource "google_cloud_scheduler_job" "health_check" {
  name     = "${var.app_name}-health-check"
  # ... scheduled health check configuration
}
```

## 📊 State Management

### Remote State

For production deployments, use remote state storage:

```hcl
# Add to main.tf
terraform {
  backend "gcs" {
    bucket  = "your-terraform-state-bucket"
    prefix  = "jgn-dev/state"
  }
}
```

Create the state bucket:

```bash
# Create state storage bucket
gsutil mb gs://your-terraform-state-bucket

# Enable versioning
gsutil versioning set on gs://your-terraform-state-bucket
```

### State Locking

Use Cloud Storage for state locking:

```hcl
terraform {
  backend "gcs" {
    bucket  = "your-terraform-state-bucket"
    prefix  = "jgn-dev/state"
    
    # Enable state locking
    lock_timeout = "5m"
  }
}
```

## 🔄 Operations

### Common Commands

```bash
# Initialize (run first time or after backend changes)
terraform init

# Plan changes (dry run)
terraform plan

# Apply changes
terraform apply

# Show current state
terraform show

# List resources
terraform state list

# Get output values
terraform output

# Format code
terraform fmt

# Validate configuration
terraform validate

# Destroy infrastructure (use with caution!)
terraform destroy
```

### Upgrading Resources

To update Cloud Run service configuration:

1. **Update variables** in `terraform.tfvars`
2. **Plan changes**: `terraform plan`
3. **Apply updates**: `terraform apply`

Example: Increase memory limit

```hcl
# In terraform.tfvars
memory_limit = "1Gi"  # Changed from 512Mi
```

### Import Existing Resources

If you have manually created resources, import them:

```bash
# Import existing Cloud Run service
terraform import google_cloud_run_v2_service.app projects/PROJECT_ID/locations/REGION/services/SERVICE_NAME

# Import service account
terraform import google_service_account.cloud_run_sa projects/PROJECT_ID/serviceAccounts/EMAIL
```

## 🚨 Troubleshooting

### Common Issues

#### 1. API Not Enabled

**Error:**
```
Error: googleapi: Error 403: Cloud Run Admin API has not been used
```

**Solution:**
```bash
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable artifactregistry.googleapis.com
```

#### 2. Insufficient Permissions

**Error:**
```
Error: googleapi: Error 403: The caller does not have permission
```

**Solution:**
Check your GCP permissions:
```bash
# Verify current user
gcloud auth list

# Check project access
gcloud projects get-iam-policy YOUR_PROJECT_ID

# Required roles:
# - roles/run.admin
# - roles/artifactregistry.admin
# - roles/cloudbuild.builds.builder
# - roles/iam.serviceAccountUser
```

#### 3. State Lock Issues

**Error:**
```
Error: Error acquiring the state lock
```

**Solution:**
```bash
# Force unlock (use carefully)
terraform force-unlock LOCK_ID

# Or wait for automatic timeout
```

#### 4. Resource Conflicts

**Error:**
```
Error: Resource already exists
```

**Solution:**
```bash
# Import existing resource
terraform import RESOURCE_TYPE.RESOURCE_NAME RESOURCE_ID

# Or rename in configuration
```

### Validation

Validate your Terraform configuration:

```bash
# Check syntax
terraform fmt -check

# Validate configuration
terraform validate

# Plan without applying
terraform plan -detailed-exitcode
```

## 🔒 Security Best Practices

### 1. Sensitive Variables

Never commit sensitive values to version control:

```hcl
# Use environment variables
export TF_VAR_github_token="your_token"

# Or use terraform.tfvars (add to .gitignore)
echo "terraform.tfvars" >> .gitignore
```

### 2. State File Security

Protect your state file:

```bash
# Use remote backend with encryption
gsutil lifecycle set lifecycle.json gs://your-terraform-state-bucket

# Restrict access to state bucket
gsutil iam ch user:user@domain.com:roles/storage.admin gs://your-terraform-state-bucket
```

### 3. Resource Tagging

Tag all resources for compliance:

```hcl
variable "labels" {
  description = "Labels to apply to resources"
  type        = map(string)
  default = {
    app         = "jgn-dev"
    environment = "production"
    managed-by  = "terraform"
    team        = "platform"
    cost-center = "engineering"
  }
}
```

## 📈 Monitoring and Alerts

### Infrastructure Monitoring

Add monitoring to your Terraform configuration:

```hcl
# Cloud Run service monitoring
resource "google_monitoring_dashboard" "cloud_run_dashboard" {
  dashboard_json = file("${path.module}/dashboards/cloud-run.json")
}

# Error rate alerting
resource "google_monitoring_alert_policy" "error_rate" {
  display_name = "High Error Rate - ${var.app_name}"
  
  conditions {
    display_name = "Cloud Run error rate"
    condition_threshold {
      filter         = "resource.type=\"cloud_run_revision\""
      comparison     = "COMPARISON_GREATER_THAN"
      threshold_value = 0.1
      duration       = "300s"
    }
  }

  notification_channels = [
    google_monitoring_notification_channel.email.name
  ]
}
```

### Cost Management

Monitor infrastructure costs:

```hcl
# Budget alert
resource "google_billing_budget" "budget" {
  billing_account = var.billing_account
  display_name    = "${var.app_name} Budget"

  budget_filter {
    projects = ["projects/${var.project_id}"]
    services = ["services/F25A61A4-397A-4FBA-9A87-2712DCDDC5B7"] # Cloud Run
  }

  amount {
    specified_amount {
      currency_code = "USD"
      units         = "50"
    }
  }

  threshold_rules {
    threshold_percent = 0.8
    spend_basis      = "CURRENT_SPEND"
  }
}
```

## 📚 Additional Resources

- [Terraform GCP Provider Documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [Cloud Run Terraform Examples](https://cloud.google.com/run/docs/configuring/services/terraform)
- [Terraform Best Practices](https://cloud.google.com/docs/terraform/best-practices)
- [GCP Resource Manager](https://cloud.google.com/resource-manager/docs)

---

**Need help?** Open an issue in the [GitHub repository](https://github.com/jgndev/jgn.dev/issues) or check the troubleshooting section above. 