#!/bin/bash

# Terraform Environment Variables Setup Script
# This script sets up environment variables that Terraform will automatically read

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Terraform environment variables...${NC}"

# Required variables
export TF_VAR_project_id="${GCP_PROJECT_ID:-}"
export TF_VAR_github_token="${GITHUB_TOKEN:-}"
export TF_VAR_github_webhook_secret="${GITHUB_WEBHOOK_SECRET:-}"

# Optional variables with defaults
export TF_VAR_region="${GCP_REGION:-us-south1}"
export TF_VAR_app_name="${APP_NAME:-jgn-dev}"
export TF_VAR_github_owner="${GITHUB_OWNER:-jgndev}"
export TF_VAR_github_repo="${GITHUB_REPO:-jgn.dev}"
export TF_VAR_custom_domain="${CUSTOM_DOMAIN:-jgn.dev}"

# Check required variables
missing_vars=()

if [ -z "$TF_VAR_project_id" ]; then
    missing_vars+=("GCP_PROJECT_ID")
fi

if [ -z "$TF_VAR_github_token" ]; then
    missing_vars+=("GITHUB_TOKEN")
fi

if [ -z "$TF_VAR_github_webhook_secret" ]; then
    missing_vars+=("GITHUB_WEBHOOK_SECRET")
fi

# Display status
if [ ${#missing_vars[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All required environment variables are set${NC}"
    echo -e "\n${YELLOW}Current configuration:${NC}"
    echo "  Project ID: $TF_VAR_project_id"
    echo "  Region: $TF_VAR_region" 
    echo "  App Name: $TF_VAR_app_name"
    echo "  GitHub Owner: $TF_VAR_github_owner"
    echo "  GitHub Repo: $TF_VAR_github_repo"
    echo "  Custom Domain: $TF_VAR_custom_domain"
    echo ""
    echo -e "${GREEN}You can now run: terraform plan or terraform apply${NC}"
else
    echo -e "\n${RED}✗ Missing required environment variables:${NC}"
    for var in "${missing_vars[@]}"; do
        echo -e "  ${RED}- $var${NC}"
    done
    echo ""
    echo -e "${YELLOW}Please set the missing variables and run this script again.${NC}"
    echo -e "${YELLOW}Example:${NC}"
    echo "  export GCP_PROJECT_ID=\"your-project-id\""
    exit 1
fi 