# GitHub Actions CI/CD Guide

This guide covers the complete CI/CD pipeline for jgn.dev using GitHub Actions. The pipeline provides automated testing, building, and deployment to GCP Cloud Run with comprehensive quality checks.

## 🎯 Overview

The CI/CD pipeline automatically:

- ✅ **Tests**: Runs Go tests, linting, and format checks
- ✅ **Builds**: Creates optimized Docker images with caching
- ✅ **Deploys**: Deploys to GCP Cloud Run on successful builds
- ✅ **Verifies**: Performs post-deployment health checks
- ✅ **Secures**: Scans for vulnerabilities on pull requests
- ✅ **Reports**: Provides detailed build and deployment status

## 🏗️ Pipeline Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Pull Request  │    │   Push to Main  │    │   Manual        │
│   Triggers      │    │   Triggers      │    │   Dispatch      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                        GitHub Actions                           │
├─────────────────┬─────────────────┬─────────────────────────────┤
│      TEST       │    SECURITY     │         DEPLOY              │
│   • Go Tests    │  • Trivy Scan   │   • Docker Build            │
│   • Linting     │  • SARIF Upload │   • Push to Registry        │
│   • Format      │  • Code Quality │   • Cloud Run Deploy        │
│   • Build       │                 │   • Health Check            │
└─────────────────┴─────────────────┴─────────────────────────────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Report   │    │  Security Alert │    │ Live Deployment │
│   Status Check  │    │  GitHub Security│    │  Service URL    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Workflow Files

The CI/CD pipeline is defined in:

```
.github/workflows/
└── deploy.yml          # Main CI/CD pipeline
```

## 🚀 Quick Setup

### Step 1: Repository Secrets

Configure these secrets in your GitHub repository (`Settings > Secrets and variables > Actions`):

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GCP_PROJECT_ID` | Your GCP project ID | `my-project-123` |
| `GCP_REGION` | GCP region for deployment | `us-central1` |
| `GCP_SA_KEY` | Service account key (JSON) | `{"type": "service_account"...}` |
| `GITHUB_TOKEN_FOR_API` | GitHub token for API access | `ghp_xxxxxxxxxxxx` |
| `GITHUB_WEBHOOK_SECRET` | Webhook secret | `your_webhook_secret` |

### Step 2: Service Account Setup

Create a GCP service account with required permissions:

```bash
# Create service account
gcloud iam service-accounts create github-actions \
    --display-name="GitHub Actions" \
    --description="Service account for GitHub Actions CI/CD"

# Get the service account email
SA_EMAIL=$(gcloud iam service-accounts list \
    --filter="displayName:GitHub Actions" \
    --format="value(email)")

# Grant required roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/artifactregistry.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/cloudbuild.builds.builder"

gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/iam.serviceAccountUser"

# Create and download key
gcloud iam service-accounts keys create key.json \
    --iam-account=$SA_EMAIL

# Add the JSON content to GitHub secrets as GCP_SA_KEY
cat key.json
```

### Step 3: Enable Required APIs

```bash
# Enable required GCP APIs
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable compute.googleapis.com
```

## 🔧 Pipeline Configuration

### Workflow Triggers

The pipeline triggers on:

```yaml
on:
  push:
    branches: [main]          # Deploy on main branch
  pull_request:
    branches: [main]          # Test on PRs to main
```

### Environment Variables

Global environment variables:

```yaml
env:
  PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
  GAR_LOCATION: ${{ secrets.GCP_REGION }}
  SERVICE: jgn-dev
  REGION: ${{ secrets.GCP_REGION }}
```

## 🧪 Jobs Breakdown

### 1. Test Job

Runs on every push and pull request:

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - name: Run tests
      run: go test -v ./...
```

**What it does:**
- Sets up Go 1.24
- Installs dependencies
- Generates Templ files
- Builds CSS with Tailwind
- Runs Go tests with race detection
- Performs linting (go vet)
- Checks code formatting (gofmt)
- Builds the application binary

### 2. Security Job

Runs vulnerability scanning on pull requests:

```yaml
security:
  runs-on: ubuntu-latest
  if: github.event_name == 'pull_request'
  steps:
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        format: 'sarif'
```

**What it does:**
- Scans filesystem for vulnerabilities
- Checks dependencies for known CVEs
- Uploads results to GitHub Security tab
- Generates SARIF report for code scanning

### 3. Deploy Job

Runs only on main branch pushes:

```yaml
deploy:
  needs: [test]
  runs-on: ubuntu-latest
  if: github.ref == 'refs/heads/main'
  steps:
    - name: Deploy to Cloud Run
      uses: google-github-actions/deploy-cloudrun@v2
```

**What it does:**
- Authenticates with GCP
- Builds Docker image with multi-stage optimization
- Pushes image to Artifact Registry
- Deploys to Cloud Run with zero-downtime
- Updates environment variables
- Provides deployment URL

### 4. Smoke Test Job

Validates deployment health:

```yaml
smoke-test:
  needs: [deploy]
  runs-on: ubuntu-latest
  steps:
    - name: Test deployment health
      run: |
        response=$(curl -s -o /dev/null -w "%{http_code}" "$SERVICE_URL/")
        if [ "$response" != "200" ]; then
          exit 1
        fi
```

**What it does:**
- Waits for deployment to stabilize
- Performs HTTP health check
- Validates response status
- Reports success/failure status

## 🔍 Monitoring and Notifications

### Build Status

The pipeline provides status updates via:

- **Commit Status**: Green/red checkmarks on commits
- **PR Checks**: Required status checks on pull requests
- **Deployment Status**: Success/failure notifications

### GitHub Integration

Status contexts created:

- `ci/test`: Test job status
- `ci/security`: Security scan status
- `deploy/cloud-run`: Deployment status
- `test/smoke`: Post-deployment health check

### Slack Notifications (Optional)

Add Slack notifications for deployment events:

```yaml
- name: Slack Notification
  if: always()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    channel: '#deployments'
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Authentication Failures

**Error:**
```
Error: google-github-actions/auth failed with: the GitHub Action workflow must specify exactly one of "workload_identity_provider" or "credentials_json"
```

**Solution:**
- Verify `GCP_SA_KEY` secret contains valid JSON
- Check service account has required permissions
- Ensure the service account key hasn't expired

#### 2. Docker Build Failures

**Error:**
```
Error: failed to solve: failed to compute cache key
```

**Solution:**
```yaml
- name: Build Docker image
  run: |
    docker build --no-cache \
      -t ${{ env.GAR_LOCATION }}-docker.pkg.dev/${{ env.PROJECT_ID }}/jgn-dev-repo/${{ env.SERVICE }}:${{ github.sha }} .
```

#### 3. Cloud Run Deployment Errors

**Error:**
```
Error: The user-provided container failed to start and listen on the port
```

**Solution:**
- Check application logs in GCP Console
- Verify container exposes port 8080
- Ensure all environment variables are set
- Check Docker image builds locally

#### 4. Permission Denied

**Error:**
```
Error: Error 403: The caller does not have permission
```

**Solution:**
- Review service account IAM roles
- Check if APIs are enabled
- Verify project ID is correct

### Debugging Steps

1. **Check Workflow Logs**:
   - Go to Actions tab in GitHub
   - Click on failed workflow run
   - Expand failed job steps

2. **Verify Secrets**:
   ```bash
   # Test GCP authentication locally
   echo $GCP_SA_KEY | base64 -d > key.json
   gcloud auth activate-service-account --key-file=key.json
   gcloud projects list
   ```

3. **Test Locally**:
   ```bash
   # Build Docker image locally
   docker build -t test-image .
   docker run -p 8080:8080 test-image
   ```

4. **Check GCP Resources**:
   ```bash
   # Verify Cloud Run service
   gcloud run services list --region=$REGION
   
   # Check Artifact Registry
   gcloud artifacts repositories list --location=$REGION
   ```

## 🔧 Customization

### Branch Protection Rules

Set up branch protection in GitHub:

1. Go to `Settings > Branches`
2. Add rule for `main` branch
3. Configure required status checks:
   - `ci/test`
   - `ci/security` (for PR branches)

### Custom Deployment Environments

Add staging environment:

```yaml
deploy-staging:
  if: github.ref == 'refs/heads/develop'
  steps:
    - name: Deploy to staging
      uses: google-github-actions/deploy-cloudrun@v2
      with:
        service: jgn-dev-staging
        region: ${{ env.REGION }}
        image: ${{ env.IMAGE_URL }}
```

### Matrix Testing

Test multiple Go versions:

```yaml
test:
  strategy:
    matrix:
      go-version: ['1.22', '1.23', '1.24']
  steps:
    - uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}
```

### Conditional Deployment

Deploy only when specific files change:

```yaml
deploy:
  if: |
    github.ref == 'refs/heads/main' &&
    (contains(github.event.head_commit.modified, 'server/') ||
     contains(github.event.head_commit.modified, 'internal/') ||
     contains(github.event.head_commit.modified, 'Dockerfile'))
```

## 📊 Performance Optimization

### Docker Build Caching

Enable Docker layer caching:

```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: ${{ steps.meta.outputs.tags }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Parallel Jobs

Run independent jobs in parallel:

```yaml
jobs:
  test:
    # ... test configuration
    
  lint:
    runs-on: ubuntu-latest
    steps:
      # ... linting steps
      
  security:
    runs-on: ubuntu-latest
    steps:
      # ... security scanning
      
  deploy:
    needs: [test, lint]  # Wait for both test and lint
```

### Artifact Caching

Cache dependencies between runs:

```yaml
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

## 🔒 Security Best Practices

### Secret Management

1. **Never log secrets**:
   ```yaml
   - name: Deploy
     run: echo "Deploying to ${{ secrets.GCP_PROJECT_ID }}"  # ❌ Don't do this
   ```

2. **Use environment variables**:
   ```yaml
   env:
     PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
   ```

3. **Rotate secrets regularly**:
   - Service account keys: Every 90 days
   - GitHub tokens: Every 6 months

### Workflow Security

1. **Restrict workflow permissions**:
   ```yaml
   permissions:
     contents: read
     id-token: write
     security-events: write
   ```

2. **Pin action versions**:
   ```yaml
   - uses: actions/checkout@v4  # ✅ Pinned version
   - uses: actions/checkout@main  # ❌ Unpinned
   ```

## 📈 Metrics and Analytics

### GitHub Actions Usage

Monitor workflow efficiency:

- **Build time**: Track job duration trends
- **Success rate**: Monitor failure rates
- **Resource usage**: Optimize runner utilization

### Deployment Metrics

Track deployment performance:

```yaml
- name: Record deployment metrics
  run: |
    echo "Deployment completed at $(date)"
    echo "Build duration: ${{ github.event.workflow_run.conclusion }}"
    echo "Commit SHA: ${{ github.sha }}"
```

## 📚 Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Google GitHub Actions](https://github.com/google-github-actions)
- [Cloud Run GitHub Actions](https://github.com/google-github-actions/deploy-cloudrun)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Trivy Action](https://github.com/aquasecurity/trivy-action)

---

**Need help?** Open an issue in the [GitHub repository](https://github.com/jgndev/jgn.dev/issues) or check the troubleshooting section above. 