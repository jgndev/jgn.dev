# Simple GCP Cloud Run Deployment

This Terraform configuration deploys your app to GCP Cloud Run with automatic GitHub integration using environment variables.

## What This Creates

- **Cloud Run Service** - Hosts your application with environment variables
- **Artifact Registry** - Stores container images  
- **Cloud Build Trigger** - Watches GitHub repo and auto-deploys on push to main
- **Custom Domain Mapping** - Maps your custom domain to the service
- **Required APIs** - Enables necessary GCP services

## Prerequisites

1. **GCP Project** with billing enabled
2. **gcloud CLI** installed and authenticated
3. **Terraform** installed
4. **GitHub repository** with your code
5. **GitHub Personal Access Token** with repo permissions

## Quick Setup

### 1. Configure Environment Variables
```bash
cd terraform
cp env.example .env
# Edit .env with your actual values
```

Required variables in `.env`:
```bash
# Required
GCP_PROJECT_ID="your-gcp-project-id"
GITHUB_TOKEN="ghp_your_github_personal_access_token"
GITHUB_WEBHOOK_SECRET="your_webhook_secret_here"

# Optional (with defaults)
GCP_REGION="us-south1"
APP_NAME="jgn-dev"
GITHUB_OWNER="jgndev"
GITHUB_REPO="jgn.dev"
CUSTOM_DOMAIN="jgn.dev"
```

### 2. Generate Required Secrets
```bash
# Generate GitHub webhook secret
openssl rand -hex 32

# Get GitHub token at: https://github.com/settings/tokens
# Required permissions: repo access
```

### 3. Deploy Infrastructure
```bash
# Load environment variables
source .env

# Validate and export Terraform variables
./setup-env.sh

# Deploy infrastructure
terraform init
terraform plan
terraform apply
```

### 4. Connect GitHub Repository (One-time)
```bash
# Connect GitHub to Google Cloud Build
gcloud builds triggers connect-github
```

### 5. Test Automatic Deployment
```bash
# Push to main branch triggers automatic deployment
git push origin main
```

## Environment Variables

Your Cloud Run service will have these environment variables set automatically:
- `PORT=8080`
- `GITHUB_TOKEN=your_token` 
- `GITHUB_WEBHOOK_SECRET=your_secret`
- `GIN_MODE=release`

## After Deployment

Your app will be available at:
- **Cloud Run URL**: `https://jgn-dev-xxxxx-uc.a.run.app` (from output)
- **Custom Domain**: `https://jgn.dev` (if configured)

## How It Works

1. **Push to main** → Cloud Build trigger activates
2. **Builds Docker image** → Pushes to Artifact Registry
3. **Deploys to Cloud Run** → Updates service with environment variables
4. **Maps custom domain** → Routes traffic to your app

## Configuration Options

All configuration is done via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GCP_PROJECT_ID` | ✅ | - | Your GCP project ID |
| `GITHUB_TOKEN` | ✅ | - | GitHub personal access token |
| `GITHUB_WEBHOOK_SECRET` | ✅ | - | Webhook signature secret |
| `GCP_REGION` | ❌ | `us-south1` | GCP deployment region |
| `APP_NAME` | ❌ | `jgn-dev` | Application name |
| `GITHUB_OWNER` | ❌ | `jgndev` | GitHub repository owner |
| `GITHUB_REPO` | ❌ | `jgn.dev` | GitHub repository name |
| `CUSTOM_DOMAIN` | ❌ | `jgn.dev` | Custom domain mapping |

## Monitoring

- **Cloud Build**: [GCP Console → Cloud Build → History](https://console.cloud.google.com/cloud-build/builds)
- **Cloud Run**: [GCP Console → Cloud Run](https://console.cloud.google.com/run)
- **Logs**: [GCP Console → Logging](https://console.cloud.google.com/logs)

## Troubleshooting

### Build Failures
```bash
# Check build logs
gcloud builds log --region=us-south1

# Manual trigger
gcloud builds triggers run jgn-dev-main-trigger
```

### Environment Variables Not Set
```bash
# Verify variables are loaded
./setup-env.sh

# Check what Terraform sees
terraform plan
```

## Cleanup

```bash
terraform destroy
``` 