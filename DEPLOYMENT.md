# StratoCraft.dev Deployment Guide

This guide will walk you through deploying the StratoCraft website to Azure with optimal cost and performance settings to achieve a 100 Lighthouse score.

## 📋 Prerequisites

Before starting, ensure you have:

- Azure CLI installed and configured
- Terraform >= 1.6.0
- Docker
- GitHub account
- Domain name (stratocraft.dev)
- Go 1.21+

## 🚀 Step 1: Initial Azure Setup

### 1.1 Create Service Principal

```bash
# Create service principal for GitHub Actions
az ad sp create-for-rbac --name "stratocraft-github-actions" \
  --role contributor \
  --scopes /subscriptions/{subscription-id} \
  --sdk-auth
```

Save the output - you'll need it for GitHub secrets.

### 1.2 Create Terraform Backend Storage

```bash
# Create resource group for Terraform state
az group create --name rg-stratocraft-tfstate --location "East US 2"

# Create storage account
az storage account create \
  --name ststratocrafttfstate \
  --resource-group rg-stratocraft-tfstate \
  --location "East US 2" \
  --sku Standard_LRS

# Create container
az storage container create \
  --name tfstate \
  --account-name ststratocrafttfstate
```

## 🏗️ Step 2: Infrastructure Deployment

### 2.1 Clone and Configure

```bash
git clone https://github.com/yourusername/stratocraft.dev.git
cd stratocraft.dev/infrastructure
```

### 2.2 Initialize Terraform

```bash
terraform init
```

### 2.3 Plan and Apply

```bash
# Review the plan
terraform plan -var="domain_name=stratocraft.dev"

# Apply the infrastructure
terraform apply -var="domain_name=stratocraft.dev"
```

### 2.4 Note Important Outputs

```bash
# Get important values for GitHub secrets
terraform output container_registry_login_server
terraform output dns_zone_name_servers
```

## 🌐 Step 3: Domain Configuration

### 3.1 Update DNS at Your Registrar

Point your domain's nameservers to the Azure DNS zone nameservers from the Terraform output:

```
ns1-01.azure-dns.com
ns2-01.azure-dns.net
ns3-01.azure-dns.org
ns4-01.azure-dns.info
```

### 3.2 Verify DNS Propagation

```bash
# Check DNS propagation
nslookup stratocraft.dev
dig stratocraft.dev
```

## 🔧 Step 4: GitHub Repository Setup

### 4.1 Fork/Create Repository

1. Fork this repository or create a new one
2. Upload all the project files

### 4.2 Configure GitHub Secrets

Add these secrets to your GitHub repository (Settings > Secrets and variables > Actions):

```bash
# Azure credentials (from Step 1.1)
AZURE_CREDENTIALS: {
  "clientId": "...",
  "clientSecret": "...",
  "subscriptionId": "...",
  "tenantId": "..."
}

# Container Registry credentials
ACR_USERNAME: [from terraform output]
ACR_PASSWORD: [from Azure portal or CLI]

# Optional: SonarCloud integration
SONAR_TOKEN: [from SonarCloud]

# Optional: Slack notifications
SLACK_WEBHOOK_URL: [from Slack]
```

### 4.3 Get ACR Credentials

```bash
# Enable admin user
az acr update --name [your-acr-name] --admin-enabled true

# Get credentials
az acr credential show --name [your-acr-name]
```

## 🚀 Step 5: First Deployment

### 5.1 Trigger GitHub Actions

Push to the main branch to trigger the CI/CD pipeline:

```bash
git add .
git commit -m "Initial deployment"
git push origin main
```

### 5.2 Monitor Deployment

1. Go to GitHub Actions tab
2. Watch the CI/CD pipeline execute
3. Check for any failures and resolve issues

### 5.3 Verify Deployment

```bash
# Get your app URL
az containerapp show \
  --name ca-stratocraft-prod \
  --resource-group rg-stratocraft-prod \
  --query properties.configuration.ingress.fqdn \
  --output tsv
```

Visit the URL to verify your application is running.

## 🔒 Step 6: SSL Certificate Setup

### 6.1 Add Custom Domain to CDN

```bash
# Add custom domain to CDN endpoint
az cdn custom-domain create \
  --endpoint-name cdn-stratocraft-prod \
  --name stratocraft-dev \
  --profile-name cdn-stratocraft-prod \
  --resource-group rg-stratocraft-prod \
  --hostname stratocraft.dev
```

### 6.2 Enable HTTPS

```bash
# Enable HTTPS with managed certificate
az cdn custom-domain enable-https \
  --endpoint-name cdn-stratocraft-prod \
  --name stratocraft-dev \
  --profile-name cdn-stratocraft-prod \
  --resource-group rg-stratocraft-prod
```

## 📊 Step 7: Performance Optimization

### 7.1 Verify Lighthouse Scores

Run Lighthouse tests to ensure 100/100 scores:

```bash
# Install Lighthouse CI
npm install -g @lhci/cli

# Run tests
lhci autorun --upload.target=temporary-public-storage
```

### 7.2 Monitor Performance

Set up monitoring dashboards in Azure:

1. Go to Application Insights
2. Create custom dashboards
3. Set up alerts for performance degradation

## 💰 Step 8: Cost Optimization

### 8.1 Verify Scale-to-Zero

Ensure your Container App scales to zero during idle periods:

```bash
# Check current replicas
az containerapp revision list \
  --name ca-stratocraft-prod \
  --resource-group rg-stratocraft-prod \
  --query '[].properties.replicas'
```

### 8.2 Monitor Costs

1. Set up Azure Cost Management alerts
2. Monitor spending in Azure portal
3. Optimize resource sizing based on usage

Expected monthly costs for this setup:
- Container Apps: $0-30 (depending on usage)
- CDN: $5-15
- DNS Zone: $0.50
- Container Registry: $5
- Application Insights: $0-10
- **Total: ~$10-60/month**

## 🔍 Step 9: Monitoring Setup

### 9.1 Configure Application Insights

```bash
# Get instrumentation key
az monitor app-insights component show \
  --app appi-stratocraft-prod \
  --resource-group rg-stratocraft-prod \
  --query instrumentationKey
```

### 9.2 Set Up Alerts

Create alerts for:
- High response times (>2 seconds)
- Error rates (>5%)
- Availability issues
- High resource usage

### 9.3 Log Analytics Queries

Useful queries for monitoring:

```kql
// Application performance
requests
| where timestamp > ago(1h)
| summarize avg(duration) by bin(timestamp, 5m)

// Error rates
exceptions
| where timestamp > ago(1h)
| summarize count() by bin(timestamp, 5m)

// Container resource usage
ContainerAppConsoleLogs_CL
| where TimeGenerated > ago(1h)
| where Log_s contains "memory" or Log_s contains "cpu"
```

## 🧪 Step 10: Testing

### 10.1 End-to-End Testing

```bash
# Test all pages
curl -I https://stratocraft.dev
curl -I https://stratocraft.dev/about
curl -I https://stratocraft.dev/blog
curl -I https://stratocraft.dev/services
curl -I https://stratocraft.dev/contact
```

### 10.2 Performance Testing

```bash
# Install Artillery
npm install -g artillery

# Run load test
artillery quick --count 10 --num 5 https://stratocraft.dev
```

### 10.3 Security Testing

```bash
# SSL Test
curl -I https://stratocraft.dev | grep -i "strict-transport-security"

# Headers test
curl -I https://stratocraft.dev | grep -i "x-frame-options\|x-content-type-options\|x-xss-protection"
```

## 🚨 Troubleshooting

### Common Issues and Solutions

#### 1. Container App Won't Start

```bash
# Check logs
az containerapp logs show \
  --name ca-stratocraft-prod \
  --resource-group rg-stratocraft-prod \
  --type console

# Check image pull
az containerapp revision list \
  --name ca-stratocraft-prod \
  --resource-group rg-stratocraft-prod
```

#### 2. DNS Issues

```bash
# Check DNS configuration
nslookup stratocraft.dev
dig @8.8.8.8 stratocraft.dev

# Check Azure DNS zone
az network dns record-set list \
  --zone-name stratocraft.dev \
  --resource-group rg-stratocraft-prod
```

#### 3. SSL Certificate Issues

```bash
# Check certificate status
az cdn custom-domain show \
  --endpoint-name cdn-stratocraft-prod \
  --name stratocraft-dev \
  --profile-name cdn-stratocraft-prod \
  --resource-group rg-stratocraft-prod
```

#### 4. GitHub Actions Failures

Common fixes:
- Verify Azure credentials in secrets
- Check ACR permissions
- Ensure Terraform state is accessible
- Verify resource names match configuration

## 🔄 Step 11: Ongoing Maintenance

### 11.1 Regular Updates

```bash
# Update dependencies monthly
go get -u ./...
go mod tidy

# Update base container image
docker pull alpine:latest
```

### 11.2 Security Updates

- Monitor for security vulnerabilities
- Update dependencies regularly
- Review and rotate secrets quarterly
- Monitor security alerts in GitHub

### 11.3 Performance Monitoring

- Weekly Lighthouse audits
- Monthly cost reviews
- Quarterly architecture reviews
- Annual disaster recovery testing

## 📈 Step 12: Scaling Considerations

### When to Scale Up

Monitor these metrics:
- Response times > 2 seconds consistently
- CPU usage > 70% sustained
- Memory usage > 80% sustained
- Error rates > 1%

### Scaling Options

1. **Vertical Scaling**: Increase CPU/memory
2. **Horizontal Scaling**: Increase max replicas
3. **Add Regions**: Deploy to multiple regions
4. **Database Scaling**: Add read replicas if using databases

## ✅ Final Checklist

Before going live, verify:

- [ ] All GitHub Actions workflows pass
- [ ] Lighthouse scores are 100/100 across all pages
- [ ] SSL certificate is properly configured
- [ ] DNS is properly configured
- [ ] Monitoring and alerts are set up
- [ ] Cost alerts are configured
- [ ] Security scans pass
- [ ] Performance tests pass
- [ ] Backup and disaster recovery plan is in place

## 📞 Support

If you encounter issues:

1. Check the [troubleshooting section](#troubleshooting)
2. Review Azure logs and metrics
3. Check GitHub Actions logs
4. Open an issue in the repository
5. Contact: hello@stratocraft.dev

---

🎉 **Congratulations!** Your high-performance, cost-optimized website is now live with enterprise-grade CI/CD and monitoring!