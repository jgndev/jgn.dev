# CI/CD Pipeline Guide

This guide covers setting up automated CI/CD pipeline for jgn.dev using both GCP Cloud Build for deployments and GitHub Actions for code quality and security.

## 🎯 Overview

The CI/CD pipeline automatically builds and deploys your application when code is pushed to the main branch:

- ✅ **Automatic Builds**: Triggered on git pushes to main branch
- ✅ **Docker Build**: Multi-stage builds with optimized layers
- ✅ **Artifact Registry**: Secure container image storage
- ✅ **Cloud Run Deployment**: Automatic deployment with zero downtime
- ✅ **Health Checks**: Built-in container health monitoring
- ✅ **Rollback Support**: Easy rollback to previous versions
- ✅ **Comprehensive CI**: GitHub Actions workflow for code quality, security, and testing on every branch and PR

## 🚦 Continuous Integration with GitHub Actions

A comprehensive GitHub Actions workflow runs on every push and pull request to any branch. It ensures code quality, security, and reliability before code is merged or deployed.

**Workflow file:** `.github/workflows/ci.yml`

### What the CI Workflow Checks

- **Formatting**: Automatically formats code with `go fmt` and organizes imports with `goimports`.
- **Static Analysis**:
  - `go vet` for common mistakes
  - `staticcheck` for advanced static analysis
  - `golangci-lint` (meta-linter, includes ineffassign, misspell, errcheck, gosec, gocyclo, deadcode, goconst, and more)
- **Vulnerability Scanning**: `govulncheck` scans for known vulnerabilities in dependencies
- **Templ Generation**: Ensures all Templ files are generated before analysis
- **Build**: Verifies the application builds successfully
- **Tests**: Runs all Go tests with race detection and coverage
- **Coverage**: Uploads coverage to Codecov for reporting

### When Does It Run?
- On every push to any branch
- On every pull request to `main`

### Example Output
- If any check fails, the workflow fails and shows the error in the GitHub Actions tab
- Formatting issues are auto-fixed (no need to fail the build for style)
- Lint, vet, and security issues must be fixed before merging

### How to Fix CI Failures
- **Formatting**: Run `go fmt ./...` and `goimports -w .`
- **Staticcheck/golangci-lint**: Read the error, fix the code, and commit
- **Vulnerabilities**: Update dependencies or address the reported issue
- **Build/Test**: Fix compilation or test failures

## 📋 Prerequisites

1. **GCP Project**: Active GCP project with billing enabled
2. **Cloud Run Service**: Deployed service (see [GCP Deployment Guide](gcp-deployment-guide.md))
3. **GitHub Repository**: Source code repository
4. **Required APIs**: Cloud Build, Cloud Run, Artifact Registry

## 🚀 Setup Cloud Build Pipeline

### Step 1: Enable Required APIs

```bash
# Enable Cloud Build API
gcloud services enable cloudbuild.googleapis.com

# Enable Cloud Run API (if not already enabled)
gcloud services enable run.googleapis.com

# Enable Artifact Registry API (if not already enabled)
gcloud services enable artifactregistry.googleapis.com
```

### Step 2: Connect GitHub Repository

1. **Open Cloud Build** in GCP Console
2. **Go to Triggers** section
3. **Click "Connect Repository"**
4. **Select GitHub** as source
5. **Authorize** GCP to access your GitHub account
6. **Select** your repository (e.g., `jgndev/jgn.dev`)
7. **Click "Connect"**

### Step 3: Create Build Trigger

#### Option A: Using GCP Console (Recommended)

1. **Navigate** to Cloud Build > Triggers
2. **Click "Create Trigger"**
3. **Configure** the following:

**Basic Settings:**
- **Name**: `jgn-dev-deploy`
- **Description**: `Deploy jgn.dev to Cloud Run`
- **Event**: Push to a branch
- **Repository**: Your connected repository
- **Branch**: `^main$` (regex for main branch)

**Configuration:**
- **Type**: Autodetected (Dockerfile)
- **Location**: Repository
- **Dockerfile**: `Dockerfile`
- **Context**: `/` (root directory)

**Advanced Settings:**
- **Service Account**: Use default Cloud Build service account
- **Timeout**: 600 seconds (10 minutes)

#### Option B: Using gcloud CLI

```bash
gcloud builds triggers create github \
    --repo-name=jgn.dev \
    --repo-owner=jgndev \
    --branch-pattern=^main$ \
    --build-config=cloudbuild.yaml \
    --description="Deploy jgn.dev to Cloud Run"
```

### Step 4: Configure Build Configuration

Create a `cloudbuild.yaml` file in your repository root:

```yaml
steps:
# Build the container image
- name: 'gcr.io/cloud-builders/docker'
  args: [
    'build',
    '-t', 'us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev:$COMMIT_SHA',
    '-t', 'us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev:latest',
    '.'
  ]

# Push the container image to Artifact Registry
- name: 'gcr.io/cloud-builders/docker'
  args: [
    'push', 
    'us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev:$COMMIT_SHA'
  ]

- name: 'gcr.io/cloud-builders/docker'
  args: [
    'push', 
    'us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev:latest'
  ]

# Deploy container image to Cloud Run
- name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
  entrypoint: gcloud
  args: [
    'run', 'deploy', 'jgn-dev',
    '--image', 'us-central1-docker.pkg.dev/$PROJECT_ID/jgn-dev-repo/jgn-dev:$COMMIT_SHA',
    '--region', 'us-central1',
    '--allow-unauthenticated',
    '--set-env-vars', 'GITHUB_TOKEN=$$GITHUB_TOKEN,GITHUB_WEBHOOK_SECRET=$$GITHUB_WEBHOOK_SECRET',
    '--cpu', '1',
    '--memory', '512Mi',
    '--min-instances', '0',
    '--max-instances', '10'
  ]
  secretEnv: ['GITHUB_TOKEN', 'GITHUB_WEBHOOK_SECRET']

# Verify deployment
- name: 'gcr.io/cloud-builders/curl'
  args: ['https://jgn.dev/']

options:
  logging: CLOUD_LOGGING_ONLY

# Configure secrets for environment variables
availableSecrets:
  secretManager:
  - versionName: projects/$PROJECT_ID/secrets/github-token/versions/latest
    env: 'GITHUB_TOKEN'
  - versionName: projects/$PROJECT_ID/secrets/github-webhook-secret/versions/latest
    env: 'GITHUB_WEBHOOK_SECRET'

# Build timeout
timeout: '600s'
```

## 🔐 Secret Management

### Step 1: Create Secrets in Secret Manager

```bash
# Enable Secret Manager API
gcloud services enable secretmanager.googleapis.com

# Create GitHub token secret
echo -n "your_github_token_here" | gcloud secrets create github-token --data-file=-

# Create webhook secret
echo -n "your_webhook_secret_here" | gcloud secrets create github-webhook-secret --data-file=-
```

### Step 2: Grant Cloud Build Access

```bash
# Get Cloud Build service account
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
CLOUD_BUILD_SA="$PROJECT_NUMBER@cloudbuild.gserviceaccount.com"

# Grant access to secrets
gcloud secrets add-iam-policy-binding github-token \
    --member="serviceAccount:$CLOUD_BUILD_SA" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding github-webhook-secret \
    --member="serviceAccount:$CLOUD_BUILD_SA" \
    --role="roles/secretmanager.secretAccessor"

# Grant Cloud Run admin role (for deployments)
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$CLOUD_BUILD_SA" \
    --role="roles/run.admin"

# Grant service account user role
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$CLOUD_BUILD_SA" \
    --role="roles/iam.serviceAccountUser"
```

## 🚀 Triggering Deployments

### Automatic Deployment

Once configured, deployments trigger automatically:

1. **Push code** to main branch
2. **Cloud Build** detects the push
3. **Builds** the Docker image
4. **Pushes** image to Artifact Registry
5. **Deploys** to Cloud Run
6. **Verifies** deployment health

### Manual Deployment

Trigger builds manually from Cloud Build console or CLI:

```bash
# Trigger build manually
gcloud builds triggers run jgn-dev-deploy --branch=main
```

## 📊 Monitoring Builds

### Cloud Build Console

1. **Navigate** to Cloud Build > History
2. **View** build logs and status
3. **Monitor** build duration and success rate
4. **Debug** failed builds

### Build Logs via CLI

```bash
# List recent builds
gcloud builds list --limit=10

# Get detailed build info
gcloud builds describe BUILD_ID

# Stream build logs
gcloud builds log BUILD_ID --stream
```

### Build Status via API

```bash
# Get build status
curl -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  "https://cloudbuild.googleapis.com/v1/projects/$PROJECT_ID/builds/BUILD_ID"
```

## 🔄 Rollback Strategy

### Automatic Rollback

Cloud Run maintains revision history:

```bash
# List revisions
gcloud run revisions list --service=jgn-dev --region=us-central1

# Rollback to previous revision
gcloud run services update-traffic jgn-dev \
    --to-revisions=REVISION_NAME=100 \
    --region=us-central1
```

### Blue-Green Deployment

For zero-downtime deployments:

```bash
# Deploy new version with traffic split
gcloud run services update-traffic jgn-dev \
    --to-revisions=NEW_REVISION=50,OLD_REVISION=50 \
    --region=us-central1

# After verification, route all traffic to new version
gcloud run services update-traffic jgn-dev \
    --to-latest \
    --region=us-central1
```

## 🚨 Troubleshooting

### Common Build Issues

1. **Permission Denied**
   ```
   Error: Permission denied accessing Secret Manager
   ```
   **Solution**: Verify Cloud Build service account has Secret Manager access

2. **Image Not Found**
   ```
   Error: Failed to pull image from Artifact Registry
   ```
   **Solution**: Check Artifact Registry repository exists and permissions

3. **Build Timeout**
   ```
   Error: Build timeout exceeded
   ```
   **Solution**: Increase timeout in `cloudbuild.yaml` or trigger configuration

4. **Cloud Run Deployment Failed**
   ```
   Error: Deployment to Cloud Run failed
   ```
   **Solution**: Check Cloud Run service exists and Cloud Build has deployment permissions

### Debug Commands

```bash
# Check Cloud Build service account
gcloud projects get-iam-policy $PROJECT_ID \
    --flatten="bindings[].members" \
    --filter="bindings.members:*@cloudbuild.gserviceaccount.com"

# Test secret access
gcloud secrets versions access latest --secret=github-token

# Check Cloud Run service
gcloud run services describe jgn-dev --region=us-central1
```

## 📈 Optimization Tips

### Build Performance

1. **Use .gcloudignore**: Exclude unnecessary files from build context
2. **Multi-stage Dockerfile**: Optimize image layers and size
3. **Build Caching**: Enable Cloud Build caching for dependencies

### Cost Optimization

1. **Efficient Triggers**: Only trigger on relevant file changes
2. **Build Machine Types**: Use appropriate machine type for build complexity
3. **Parallel Builds**: Avoid unnecessary concurrent builds

### Security Best Practices

1. **Least Privilege**: Grant minimal required permissions
2. **Secret Rotation**: Regularly rotate secrets
3. **Audit Logs**: Monitor Cloud Build audit logs
4. **Private Images**: Use private Artifact Registry repositories

## 📋 Maintenance

### Regular Tasks

1. **Monitor build success rate**
2. **Review build logs for warnings**
3. **Update secrets when GitHub tokens expire**
4. **Clean up old container images**
5. **Review and update build configurations**

### Backup Strategy

1. **Source Code**: Stored in GitHub (primary backup)
2. **Container Images**: Stored in Artifact Registry with retention policy
3. **Build History**: Maintained by Cloud Build (90 days default)
4. **Configuration**: Version controlled in repository

---

🔗 **Related Guides:**
- [GCP Deployment Guide](gcp-deployment-guide.md)
- [Webhook Setup Guide](webhook-setup-guide.md)
- [GitHub Token Setup Guide](github-token-setup-guide.md) 