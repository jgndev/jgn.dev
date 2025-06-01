#!/bin/bash

# Deploy jgn.dev to GCP Cloud Run
# This script handles the initial setup and deployment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID=${PROJECT_ID:-""}
REGION=${REGION:-"us-central1"}
APP_NAME=${APP_NAME:-"jgn-dev"}
GITHUB_TOKEN=${GITHUB_TOKEN:-""}
GITHUB_WEBHOOK_SECRET=${GITHUB_WEBHOOK_SECRET:-""}

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if gcloud is installed
    if ! command -v gcloud &> /dev/null; then
        print_error "gcloud CLI is not installed. Please install it from: https://cloud.google.com/sdk/docs/install"
        exit 1
    fi

    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install it from: https://docs.docker.com/get-docker/"
        exit 1
    fi

    # Check if Terraform is installed
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install it from: https://www.terraform.io/downloads"
        exit 1
    fi

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install it from: https://golang.org/dl/"
        exit 1
    fi

    # Check if Node.js is installed
    if ! command -v node &> /dev/null; then
        print_error "Node.js is not installed. Please install it from: https://nodejs.org/"
        exit 1
    fi

    print_success "All prerequisites are installed"
}

# Validate environment variables
validate_environment() {
    print_status "Validating environment variables..."
    
    if [[ -z "$PROJECT_ID" ]]; then
        print_error "PROJECT_ID environment variable is not set"
        print_error "Please set it with: export PROJECT_ID=your-gcp-project-id"
        exit 1
    fi

    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_error "GITHUB_TOKEN environment variable is not set"
        print_error "Please set it with: export GITHUB_TOKEN=your_github_token"
        print_error "See docs/github-token-setup-guide.md for instructions"
        exit 1
    fi

    print_success "Environment variables validated"
}

# Set up GCP authentication
setup_gcp_auth() {
    print_status "Setting up GCP authentication..."
    
    # Check if user is authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "."; then
        print_warning "Not authenticated with GCP. Starting authentication..."
        gcloud auth login
    fi

    # Set the project
    gcloud config set project "$PROJECT_ID"
    
    # Enable required APIs
    print_status "Enabling required GCP APIs..."
    gcloud services enable run.googleapis.com
    gcloud services enable cloudbuild.googleapis.com
    gcloud services enable artifactregistry.googleapis.com
    
    print_success "GCP authentication and APIs configured"
}

# Build and test locally
build_and_test() {
    print_status "Building and testing the application locally..."
    
    # Install Go dependencies
    go mod download
    
    # Install templ if not already installed
    if ! command -v templ &> /dev/null; then
        print_status "Installing templ CLI..."
        go install github.com/a-h/templ/cmd/templ@latest
    fi

    # Install Node.js dependencies
    npm ci

    # Generate templ files
    templ generate

    # Build CSS
    npx tailwindcss -i ./public/css/style.css -o ./public/css/site.css --minify

    # Run tests
    print_status "Running tests..."
    go test -v ./...

    # Run linting
    go vet ./...

    print_success "Build and tests completed successfully"
}

# Deploy with Terraform
deploy_with_terraform() {
    print_status "Deploying infrastructure with Terraform..."
    
    cd terraform

    # Check if terraform.tfvars exists
    if [[ ! -f "terraform.tfvars" ]]; then
        print_warning "terraform.tfvars not found. Creating from example..."
        cp terraform.tfvars.example terraform.tfvars
        
        # Update with current values
        sed -i.bak "s/your-gcp-project-id/$PROJECT_ID/g" terraform.tfvars
        sed -i.bak "s/ghp_your_github_personal_access_token/$GITHUB_TOKEN/g" terraform.tfvars
        
        if [[ -n "$GITHUB_WEBHOOK_SECRET" ]]; then
            sed -i.bak "s/your_webhook_secret_here/$GITHUB_WEBHOOK_SECRET/g" terraform.tfvars
        fi
        
        rm terraform.tfvars.bak
        
        print_warning "Please review and update terraform.tfvars before continuing"
        print_warning "Press Enter to continue or Ctrl+C to abort"
        read -r
    fi

    # Initialize Terraform
    terraform init

    # Plan deployment
    terraform plan

    # Ask for confirmation
    print_warning "Review the Terraform plan above. Do you want to proceed with deployment? (y/N)"
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        print_error "Deployment aborted by user"
        exit 1
    fi

    # Apply deployment
    terraform apply -auto-approve

    cd ..
    print_success "Infrastructure deployed successfully"
}

# Build and push Docker image
build_and_push_image() {
    print_status "Building and pushing Docker image..."
    
    # Configure Docker to use gcloud as credential helper
    gcloud auth configure-docker "${REGION}-docker.pkg.dev"

    # Build the image
    IMAGE_URL="${REGION}-docker.pkg.dev/${PROJECT_ID}/${APP_NAME}-repo/${APP_NAME}:latest"
    docker build -t "$IMAGE_URL" .

    # Push the image
    docker push "$IMAGE_URL"

    print_success "Docker image built and pushed successfully"
}

# Deploy to Cloud Run
deploy_to_cloud_run() {
    print_status "Deploying to Cloud Run..."
    
    IMAGE_URL="${REGION}-docker.pkg.dev/${PROJECT_ID}/${APP_NAME}-repo/${APP_NAME}:latest"
    
    gcloud run deploy "$APP_NAME" \
        --image="$IMAGE_URL" \
        --region="$REGION" \
        --allow-unauthenticated \
        --set-env-vars="GITHUB_TOKEN=$GITHUB_TOKEN,GITHUB_WEBHOOK_SECRET=$GITHUB_WEBHOOK_SECRET,PORT=8080,GIN_MODE=release" \
        --cpu=1 \
        --memory=512Mi \
        --min-instances=0 \
        --max-instances=10 \
        --platform=managed

    # Get the service URL
    SERVICE_URL=$(gcloud run services describe "$APP_NAME" --region="$REGION" --format="value(status.url)")
    
    print_success "Application deployed successfully!"
    print_success "Service URL: $SERVICE_URL"
    print_success "Webhook URL: $SERVICE_URL/webhook/github"
}

# Main deployment function
main() {
    echo -e "${GREEN}"
    echo "=================================================="
    echo "     jgn.dev GCP Cloud Run Deployment Script     "
    echo "=================================================="
    echo -e "${NC}"

    check_prerequisites
    validate_environment
    setup_gcp_auth
    build_and_test
    deploy_with_terraform
    build_and_push_image
    deploy_to_cloud_run

    echo -e "${GREEN}"
    echo "=================================================="
    echo "           Deployment Completed Successfully!    "
    echo "=================================================="
    echo -e "${NC}"
    
    print_status "Next steps:"
    echo "1. Update your domain DNS to point to the Cloud Run service"
    echo "2. Configure GitHub webhook with the webhook URL shown above"
    echo "3. Test the deployment by visiting your service URL"
    echo "4. Set up monitoring and alerts in GCP Console"
}

# Handle script arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  help    Show this help message"
        echo "  check   Check prerequisites only"
        echo ""
        echo "Environment variables:"
        echo "  PROJECT_ID              GCP project ID (required)"
        echo "  GITHUB_TOKEN           GitHub personal access token (required)"
        echo "  GITHUB_WEBHOOK_SECRET  GitHub webhook secret (optional)"
        echo "  REGION                 GCP region (default: us-central1)"
        echo "  APP_NAME               Application name (default: jgn-dev)"
        ;;
    "check")
        check_prerequisites
        validate_environment
        ;;
    *)
        main
        ;;
esac 