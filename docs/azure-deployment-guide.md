# Azure App Service Deployment Guide

This guide covers deploying stratocraft.dev to Azure App Service for Containers, providing the best web application experience with built-in SSL, domain management, and continuous deployment.

## Why Azure App Service for Containers?

Azure App Service is specifically designed for web applications and offers several advantages over other container services:

- **🌐 Web-First Design**: Built specifically for web applications
- **🔒 Built-in SSL**: Automatic HTTPS certificates and domain management  
- **⚡ Always-On**: No cold starts, faster response times
- **💰 Better Value**: B1 tier (~$13-15/month) with comprehensive features
- **🔄 CI/CD Integration**: Automatic deployment when you push new images
- **📊 Monitoring**: Built-in Application Insights and health monitoring
- **🌍 Custom Domains**: Easy setup for your own domain

## Prerequisites

Before deploying, ensure you have:

1. **Azure CLI** installed and configured
2. **Docker** installed and running
3. **GitHub Personal Access Token** - [Setup Guide](github-token-setup-guide.md)
4. **Azure Subscription** with appropriate permissions

## Quick Deployment

### 1. Set Environment Variables

```bash
export GITHUB_TOKEN=your_github_personal_access_token
export GITHUB_WEBHOOK_SECRET=$(openssl rand -hex 32)  # Optional
```

### 2. Run Deployment Script

```bash
./scripts/deploy-azure-appservice.sh
```

That's it! The script handles everything automatically.

## Manual Deployment Steps

If you prefer manual control or need to customize the deployment:

### 1. Build and Push Container Image

```bash
# Create builder for AMD64 architecture
docker buildx create --use --name appservice-builder

# Build for Linux/AMD64 (required for Azure)
docker buildx build --platform linux/amd64 -t stratocraft-dev:latest --load .

# Create Azure Container Registry (if needed)
az acr create --resource-group stratocraft-rg --name stratocraftacr --sku Basic

# Tag and push image
az acr login --name stratocraftacr
docker tag stratocraft-dev:latest stratocraftacr.azurecr.io/stratocraft-dev:latest
docker push stratocraftacr.azurecr.io/stratocraft-dev:latest
```

### 2. Create App Service Plan

```bash
# Create B1 tier App Service Plan
az appservice plan create \
    --name stratocraft-plan \
    --resource-group stratocraft-rg \
    --location centralus \
    --sku B1 \
    --is-linux
```

### 3. Create and Configure Web App

```bash
# Create the web app
az webapp create \
    --resource-group stratocraft-rg \
    --plan stratocraft-plan \
    --name stratocraft-webapp \
    --container-image-name stratocraftacr.azurecr.io/stratocraft-dev:latest

# Configure container settings
az webapp config container set \
    --name stratocraft-webapp \
    --resource-group stratocraft-rg \
    --container-image-name stratocraftacr.azurecr.io/stratocraft-dev:latest \
    --container-registry-url https://stratocraftacr.azurecr.io \
    --container-registry-user stratocraftacr \
    --container-registry-password $(az acr credential show --name stratocraftacr --query passwords[0].value --output tsv)

# Set environment variables
az webapp config appsettings set \
    --resource-group stratocraft-rg \
    --name stratocraft-webapp \
    --settings GITHUB_TOKEN=$GITHUB_TOKEN GITHUB_WEBHOOK_SECRET=$GITHUB_WEBHOOK_SECRET WEBSITES_PORT=8080
```

## Cost Breakdown

### App Service B1 Tier (~$13-15/month)

| Component | Specification | Monthly Cost |
|-----------|---------------|--------------|
| Compute | 1 vCPU, 1.75 GB RAM | ~$13.00 |
| Storage | 10 GB included | $0.00 |
| Bandwidth | 5 GB free, then $0.087/GB | ~$0-2.00 |
| Container Registry | Basic tier | ~$5.00 |
| **Total** | | **~$18-20/month** |

### Cost Optimization Tips

1. **Use B1 tier** for development/small sites
2. **Monitor bandwidth** usage to avoid overage charges
3. **Enable auto-scaling** only if needed (adds cost)
4. **Use Azure CDN** for static content caching
5. **Set up alerts** for cost monitoring

## Configuration Options

### Custom Domain Setup

```bash
# Add custom domain
az webapp config hostname add \
    --resource-group stratocraft-rg \
    --webapp-name stratocraft-webapp \
    --hostname yourdomain.com

# Enable SSL (free with App Service managed certificate)
az webapp config ssl bind \
    --resource-group stratocraft-rg \
    --name stratocraft-webapp \
    --certificate-thumbprint auto \
    --ssl-type SNI
```

### Continuous Deployment

```bash
# Enable continuous deployment (auto-pull new images)
az webapp deployment container config \
    --resource-group stratocraft-rg \
    --name stratocraft-webapp \
    --enable-cd true
```

### Application Insights

```bash
# Create Application Insights
az monitor app-insights component create \
    --app stratocraft-insights \
    --location centralus \
    --resource-group stratocraft-rg

# Get instrumentation key
INSTRUMENTATION_KEY=$(az monitor app-insights component show \
    --app stratocraft-insights \
    --resource-group stratocraft-rg \
    --query instrumentationKey --output tsv)

# Add to app settings
az webapp config appsettings set \
    --resource-group stratocraft-rg \
    --name stratocraft-webapp \
    --settings APPINSIGHTS_INSTRUMENTATIONKEY=$INSTRUMENTATION_KEY
```

## Management Commands

### Monitoring and Logs

```bash
# View live logs
az webapp log tail --resource-group stratocraft-rg --name stratocraft-webapp

# Download log files
az webapp log download --resource-group stratocraft-rg --name stratocraft-webapp

# Check app status
az webapp show --resource-group stratocraft-rg --name stratocraft-webapp --query state
```

### Scaling and Updates

```bash
# Scale up to S1 tier (more performance)
az appservice plan update \
    --resource-group stratocraft-rg \
    --name stratocraft-plan \
    --sku S1

# Restart application
az webapp restart --resource-group stratocraft-rg --name stratocraft-webapp

# Update container image
az webapp config container set \
    --name stratocraft-webapp \
    --resource-group stratocraft-rg \
    --container-image-name stratocraftacr.azurecr.io/stratocraft-dev:latest
```

## Troubleshooting

### Common Issues

**App won't start:**
```bash
# Check application logs
az webapp log tail --resource-group stratocraft-rg --name stratocraft-webapp

# Verify port configuration
az webapp config show --resource-group stratocraft-rg --name stratocraft-webapp --query 'siteConfig.appSettings'
```

**Container pull fails:**
```bash
# Verify registry credentials
az acr credential show --name stratocraftacr

# Test registry connectivity
az acr login --name stratocraftacr
docker pull stratocraftacr.azurecr.io/stratocraft-dev:latest
```

**Environment variables not working:**
```bash
# List current app settings
az webapp config appsettings list --resource-group stratocraft-rg --name stratocraft-webapp

# Update specific setting
az webapp config appsettings set \
    --resource-group stratocraft-rg \
    --name stratocraft-webapp \
    --settings GITHUB_TOKEN=new_token
```

### Performance Optimization

1. **Enable CDN** for static content
2. **Use Application Insights** for monitoring
3. **Configure health checks** for auto-restart
4. **Set up staging slots** for zero-downtime deployments

## Security Considerations

1. **Use managed identity** instead of registry passwords when possible
2. **Enable HTTPS only** in App Service settings
3. **Configure IP restrictions** if needed
4. **Regularly rotate** GitHub tokens and webhook secrets
5. **Monitor access logs** for suspicious activity

## Next Steps

After deployment:

1. **Configure custom domain** and SSL certificate
2. **Set up Azure CDN** for better performance
3. **Enable Application Insights** for monitoring
4. **Configure webhooks** for automatic deployments
5. **Set up backup and disaster recovery**

For webhook configuration, see the [Webhook Setup Guide](webhook-setup-guide.md).