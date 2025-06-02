# Required: Your GCP Project ID
project_id = "jgn-dev-454618"

# Optional: GCP region (default: us-central1)
region = "us-south1"

# Optional: Application name (default: jgn-dev)
app_name = "jgn-dev"

# Optional: GitHub repository details
github_owner = "jgndev"
github_repo  = "jgn.dev"

# Optional: Custom Domain
# custom_domain = "jgn.dev"

# Resource Limits (optional - defaults shown)
# min_instances = 0
# max_instances = 10
# cpu_limit     = "1"
# memory_limit  = "512Mi"

# Environment
environment = "prod"

# Labels
labels = {
  app         = "jgn-dev"
  environment = "production"
  managed-by  = "terraform"
  owner       = "jgndev"
} 