# Deployment Setup

## Overview

This project uses GitHub Actions for CI/CD with the following workflows:

- **CI Pipeline** (`ci.yml`): Runs on all branches except `main`, validates code quality, formatting, and runs tests
- **Deployment Pipeline** (`deploy.yml`): Runs only when PRs are merged to `main`, builds Docker images and deploys to Cloud Run

## Required GitHub Secrets

You need to configure the following secrets in your GitHub repository settings:

### 1. `GCP_PROJECT_ID`
Your Google Cloud Project ID (e.g., `my-project-123`)

### 2. `GCP_SA_KEY` 
Service Account Key JSON with the following permissions:
- Cloud Run Admin
- Artifact Registry Writer
- Storage Admin
- Service Account User

To create this:
1. Go to GCP Console → IAM & Admin → Service Accounts
2. Create a new service account or use existing one
3. Add the required roles listed above
4. Generate a JSON key
5. Copy the entire JSON content and add it as a GitHub secret

### 3. `GH_TOKEN`
Your GitHub Personal Access Token (used by the application)

> **Note**: GitHub doesn't allow secrets starting with `GITHUB_`, so we use `GH_TOKEN` instead.

### 4. `WEBHOOK_SECRET`
Your GitHub Webhook Secret (used by the application)

> **Note**: GitHub doesn't allow secrets starting with `GITHUB_`, so we use `WEBHOOK_SECRET` instead.

## Service Account Setup

```bash
# Create service account
gcloud iam service-accounts create github-actions \
    --display-name="GitHub Actions"

# Add required roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/storage.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/iam.serviceAccountUser"

# Generate key
gcloud iam service-accounts keys create key.json \
    --iam-account=github-actions@$PROJECT_ID.iam.gserviceaccount.com
```

## Artifact Registry Setup

Make sure your Artifact Registry repository exists:

```bash
gcloud artifacts repositories create jgn-dev-repo \
    --repository-format=docker \
    --location=us-central1 \
    --description="Docker repository for jgn.dev"
```

## GitHub Secrets Configuration

### Step-by-Step Setup

1. **Go to your GitHub repository**
2. **Navigate to**: Settings → Secrets and variables → Actions
3. **Click**: "New repository secret"
4. **Add each secret**:

| Secret Name      | Value                  | Description                  |
|------------------|------------------------|------------------------------|
| `GCP_PROJECT_ID` | `your-project-id`      | Your GCP project ID          |
| `GCP_SA_KEY`     | `{...json content...}` | Service account JSON key     |
| `GH_TOKEN`       | `ghp_xxxxxxxxxxxx`     | GitHub Personal Access Token |
| `WEBHOOK_SECRET` | `your-webhook-secret`  | GitHub webhook secret        |

### Environment Variable Mapping

The deployment workflow maps these secrets to environment variables in Cloud Run:

```yaml
# GitHub Secret → Cloud Run Environment Variable
GH_TOKEN → GITHUB_TOKEN
WEBHOOK_SECRET → GITHUB_WEBHOOK_SECRET
```

This maintains compatibility with your existing application code while avoiding GitHub's secret naming restrictions.

## Workflow Benefits

### Cost Optimization
- **Before**: Cloud Build pulls source, builds, and deploys (slower, more expensive)
- **After**: GitHub Actions builds once, pushes image, Cloud Run uses pre-built image (faster, cheaper)

### Performance Improvements
- Pre-built Docker images deploy much faster
- No build time during deployment = faster rollbacks
- Cached layers reduce build times on GitHub Actions

### Additional Features
- Automatic image cleanup (keeps only 10 most recent)
- Deployment verification with health checks
- Proper tagging with commit SHA and latest
- Build metadata injection

## Migration from Cloud Build

1. **Set up GitHub secrets** as documented above
2. **Disable your current Cloud Build triggers**:
   ```bash
   # List existing triggers
   gcloud builds triggers list
   
   # Delete triggers (replace TRIGGER_ID with actual ID)
   gcloud builds triggers delete TRIGGER_ID
   ```
3. **Test the new pipeline** by creating and merging a PR to main
4. **Verify the deployment** works correctly
5. **Remove the old `cloudbuild.yaml`** file (optional)

## Troubleshooting

### Common Issues

1. **Authentication errors**: Verify `GCP_SA_KEY` is valid JSON and has correct permissions
2. **Image push failures**: Check Artifact Registry repository exists and SA has write access
3. **Deploy failures**: Verify Cloud Run service name and region match the workflow
4. **Health check failures**: Ensure your app responds on port 8080 and path `/`
5. **Secret naming errors**: Remember GitHub secrets cannot start with `GITHUB_`

### Debug Commands

```bash
# Check service account permissions
gcloud projects get-iam-policy $PROJECT_ID \
    --flatten="bindings[].members" \
    --filter="bindings.members:github-actions@$PROJECT_ID.iam.gserviceaccount.com"

# List artifacts
gcloud artifacts docker images list us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev

# Check Cloud Run service
gcloud run services describe jgn-dev --region=us-central1

# Test GitHub secrets (in Actions workflow)
echo "GH_TOKEN is set: ${{ secrets.GH_TOKEN != '' }}"
echo "WEBHOOK_SECRET is set: ${{ secrets.WEBHOOK_SECRET != '' }}"
```

## Security Best Practices

1. **Rotate secrets regularly**: Update GitHub tokens and webhook secrets periodically
2. **Least privilege**: Service account should only have required permissions
3. **Monitor access**: Review service account usage in GCP Console
4. **Secure webhooks**: Always use webhook secrets for validation

## Testing the Setup

### Test CI Pipeline
1. Create a feature branch
2. Make a small change and push
3. Verify CI workflow runs and passes all checks

### Test Deployment Pipeline
1. Create and merge a PR to main
2. Watch the deployment workflow in GitHub Actions
3. Verify the new image is pushed to Artifact Registry
4. Confirm Cloud Run service is updated and healthy

### Verify Environment Variables
Check that your application receives the correct environment variables:
```bash
# In your Cloud Run service logs, you should see:
# GITHUB_TOKEN=ghp_xxxxxxxxxxxx (from GH_TOKEN secret)
# GITHUB_WEBHOOK_SECRET=your-secret (from WEBHOOK_SECRET secret)
``` 