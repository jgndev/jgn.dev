# GCP Cloud Run Deployment Guide

This guide walks you through deploying jgn.dev to Google Cloud Platform (GCP) Cloud Run, a serverless container platform that automatically scales your application based on traffic.

## 🎯 Overview

**Cloud Run Benefits:**
- ✅ **Pay-per-use**: Only pay when requests are being handled
- ✅ **Auto-scaling**: Scales to zero when not in use, up to 1000 instances
- ✅ **Serverless**: No server management required
- ✅ **Fast deployments**: Containers start in seconds
- ✅ **Built-in CI/CD**: Integration with Cloud Build and GitHub
- ✅ **Global**: Deploy to multiple regions worldwide

## 📋 Prerequisites

1. **GCP Account**: Create at [cloud.google.com](https://cloud.google.com)
2. **GCP Project**: Create a new project or use existing
3. **Billing Enabled**: Required for Cloud Run (free tier available)
4. **gcloud CLI**: Install from [cloud.google.com/sdk](https://cloud.google.com/sdk)
5. **Docker**: Install from [docker.com](https://docker.com)
6. **Terraform**: Install from [terraform.io](https://terraform.io)

## 🚀 Quick Deployment

### Option 1: Automated Script (Recommended)

```bash
# Set required environment variables
export PROJECT_ID=your-gcp-project-id
export GITHUB_TOKEN=your_github_token
export GITHUB_WEBHOOK_SECRET=$(openssl rand -hex 32)

# Run the deployment script
./scripts/deploy-gcp-cloud-run.sh
```

### Option 2: Manual Deployment

Follow these steps for a manual deployment:

#### Step 1: GCP Setup

```bash
# Install gcloud CLI (if not installed)
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Authenticate with GCP
gcloud auth login

# Set your project
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
gcloud services enable artifactregistry.googleapis.com
```

#### Step 2: Infrastructure with Terraform

```bash
# Navigate to terraform directory
cd terraform

# Copy and configure variables
cp terraform.tfvars.example terraform.tfvars

# Edit terraform.tfvars with your values
nano terraform.tfvars

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the infrastructure
terraform apply
```

#### Step 3: Build and Deploy

```bash
# Configure Docker for Artifact Registry
gcloud auth configure-docker us-central1-docker.pkg.dev

# Build the container image
docker build -t us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:latest .

# Push the image
docker push us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:latest

# Deploy to Cloud Run
gcloud run deploy jgn-dev \
    --image=us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:latest \
    --region=us-central1 \
    --allow-unauthenticated \
    --set-env-vars=GITHUB_TOKEN=your_token,GITHUB_WEBHOOK_SECRET=your_secret \
    --cpu=1 \
    --memory=512Mi \
    --min-instances=0 \
    --max-instances=10
```

## 🔧 Configuration

### Environment Variables

Configure these in your Cloud Run service:

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub personal access token for API access |
| `GITHUB_WEBHOOK_SECRET` | No | Secret for webhook signature verification |
| `PORT` | No | Server port (default: 8080) |
| `GIN_MODE` | No | Gin framework mode (default: release) |

### Resource Configuration

Recommended settings for different traffic levels:

#### Low Traffic (Personal Blog)
```bash
--cpu=1
--memory=512Mi
--min-instances=0
--max-instances=5
```

#### Medium Traffic (Company Blog)
```bash
--cpu=2
--memory=1Gi
--min-instances=1
--max-instances=20
```

#### High Traffic (Popular Site)
```bash
--cpu=4
--memory=2Gi
--min-instances=3
--max-instances=100
```

## 🌐 Custom Domain Setup

### Step 1: Domain Mapping

```bash
# Map custom domain to Cloud Run service
gcloud run domain-mappings create \
    --service=jgn-dev \
    --domain=yourdomain.com \
    --region=us-central1
```

### Step 2: DNS Configuration

Add these DNS records to your domain provider:

1. **CNAME Record**: 
   - Name: `www` (or `@` for apex domain)
   - Value: `ghs.googlehosted.com`

2. **Verification TXT Record** (if required):
   - Name: `@`
   - Value: (provided by GCP Console)

### Step 3: SSL Certificate

Cloud Run automatically provisions SSL certificates for custom domains. This typically takes 15-30 minutes.

## 📊 Monitoring and Logging

### Cloud Run Metrics

Access metrics in the GCP Console under Cloud Run > Services > jgn-dev:

- **Request count**: Number of requests served
- **Request latency**: Response time distribution
- **Container instances**: Active instance count
- **CPU utilization**: Container CPU usage
- **Memory utilization**: Container memory usage

### Application Logs

View application logs:

```bash
# View recent logs
gcloud logs read --filter="resource.type=cloud_run_revision" --limit=50

# Follow logs in real-time
gcloud logs tail --filter="resource.type=cloud_run_revision"
```

### Alerts Setup

Set up monitoring alerts:

1. Go to **Monitoring** in GCP Console
2. Create **Alerting Policies**
3. Configure notifications for:
   - High error rates (>5%)
   - High latency (>2 seconds)
   - Memory usage (>80%)

## 💰 Cost Optimization

### Pricing Model

Cloud Run uses a pay-per-use model:

- **CPU**: $0.00002400 per vCPU-second
- **Memory**: $0.00000250 per GiB-second
- **Requests**: $0.40 per million requests
- **Free Tier**: 2 million requests/month, 400,000 GiB-seconds/month

### Cost Examples

**Personal Blog (1,000 visits/month):**
- Requests: 1,000 × 3 requests/visit = 3,000 requests
- CPU time: 3,000 × 0.1 seconds = 300 CPU-seconds
- **Estimated cost: $0.50-2.00/month**

**Medium Blog (10,000 visits/month):**
- Requests: 10,000 × 3 requests/visit = 30,000 requests
- CPU time: 30,000 × 0.1 seconds = 3,000 CPU-seconds
- **Estimated cost: $3-8/month**

### Optimization Tips

1. **Right-size resources**: Start with minimal CPU/memory
2. **Use min-instances=0**: Allow scaling to zero
3. **Optimize container**: Use multi-stage Docker builds
4. **Enable request caching**: Reduce backend calls
5. **Monitor usage**: Review metrics monthly

## 🔄 CI/CD Integration

### GitHub Actions

The deployment includes automatic CI/CD via GitHub Actions:

1. **Trigger**: Push to main branch
2. **Test**: Run Go tests and linting
3. **Build**: Create Docker image
4. **Deploy**: Update Cloud Run service
5. **Verify**: Health check deployment

### Manual Deployment

For manual deployments:

```bash
# Build and push new image
docker build -t us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:v1.2.3 .
docker push us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:v1.2.3

# Update Cloud Run service
gcloud run services update jgn-dev \
    --image=us-central1-docker.pkg.dev/YOUR_PROJECT_ID/jgn-dev-repo/jgn-dev:v1.2.3 \
    --region=us-central1
```

## 🔒 Security

### Service Account

The deployment creates a dedicated service account with minimal permissions:

- `roles/run.invoker`: Allow service invocation
- Custom IAM bindings for specific resources

### Network Security

- **HTTPS only**: All traffic encrypted in transit
- **VPC connector**: Optional private network access
- **Identity-Aware Proxy**: Optional user authentication

### Secret Management

Store sensitive data in Google Secret Manager:

```bash
# Create secret
gcloud secrets create github-token --data-file=token.txt

# Grant access to Cloud Run service
gcloud secrets add-iam-policy-binding github-token \
    --member="serviceAccount:jgn-dev-run-sa@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Build Failures

**Problem**: Docker build fails
```bash
ERROR: failed to solve: failed to compute cache key
```

**Solution**: 
```bash
# Clear Docker cache
docker system prune -a

# Rebuild image
docker build --no-cache -t your-image .
```

#### 2. Deployment Errors

**Problem**: Cloud Run deployment fails
```bash
ERROR: The user-provided container failed to start and listen on the port
```

**Solution**:
- Verify your app listens on port 8080
- Check container logs for errors
- Ensure all dependencies are included

#### 3. Domain Mapping Issues

**Problem**: Custom domain not working
```bash
Failed to verify domain ownership
```

**Solution**:
- Add verification TXT record to DNS
- Wait 15-30 minutes for propagation
- Verify DNS settings with `nslookup`

#### 4. Performance Issues

**Problem**: Slow response times
```bash
High latency in Cloud Run metrics
```

**Solution**:
- Increase CPU allocation
- Add min-instances to avoid cold starts
- Optimize application code
- Enable caching

### Getting Help

1. **GCP Console**: Check service logs and metrics
2. **Support**: Create support case if needed
3. **Community**: Stack Overflow with `google-cloud-run` tag
4. **Documentation**: [cloud.google.com/run/docs](https://cloud.google.com/run/docs)

## ✅ Next Steps

After successful deployment:

1. **Configure monitoring**: Set up alerts and dashboards
2. **Setup backup**: Configure automated backups if needed
3. **Performance tuning**: Monitor and optimize based on usage
4. **Security review**: Implement additional security measures
5. **Documentation**: Update team documentation

## 📚 Additional Resources

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Cloud Run Pricing](https://cloud.google.com/run/pricing)
- [Best Practices](https://cloud.google.com/run/docs/best-practices)
- [Security Guide](https://cloud.google.com/run/docs/securing)
- [Monitoring Guide](https://cloud.google.com/run/docs/monitoring)

---

**Need help?** Open an issue in the [GitHub repository](https://github.com/jgndev/jgn.dev/issues) or check the troubleshooting section above. 