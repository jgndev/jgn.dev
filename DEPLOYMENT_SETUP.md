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

### 3. `GITHUB_TOKEN`
Your GitHub Personal Access Token (used by the application)

### 4. `GITHUB_WEBHOOK_SECRET`
Your GitHub Webhook Secret (used by the application)

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

1. Disable your current Cloud Build triggers
2. Set up the GitHub secrets listed above
3. Merge a PR to main to trigger the new deployment pipeline
4. Verify the deployment works correctly
5. Remove the old `cloudbuild.yaml` file (optional)

## Troubleshooting

### Common Issues

1. **Authentication errors**: Verify `GCP_SA_KEY` is valid JSON and has correct permissions
2. **Image push failures**: Check Artifact Registry repository exists and SA has write access
3. **Deploy failures**: Verify Cloud Run service name and region match the workflow
4. **Health check failures**: Ensure your app responds on port 8080 and path `/`

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
``` 