#!/bin/bash

# Deploy jgn.dev to GCP Cloud Run
# This script handles manual deployment and initial setup for Cloud Build CI/CD

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
REPO_NAME=${REPO_NAME:-"jgn-dev-repo"}
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
    gcloud services enable secretmanager.googleapis.com
    
    print_success "GCP authentication and APIs configured"
}

# Create Artifact Registry repository
create_artifact_registry() {
    print_status "Setting up Artifact Registry..."
    
    # Check if repository exists
    if gcloud artifacts repositories describe "$REPO_NAME" --location="$REGION" &>/dev/null; then
        print_success "Artifact Registry repository '$REPO_NAME' already exists"
    else
        print_status "Creating Artifact Registry repository..."
        gcloud artifacts repositories create "$REPO_NAME" \
            --repository-format=docker \
            --location="$REGION" \
            --description="Docker repository for jgn.dev"
        print_success "Artifact Registry repository created"
    fi
    
    # Configure Docker authentication
    gcloud auth configure-docker "${REGION}-docker.pkg.dev"
}

# Set up secrets in Secret Manager
setup_secrets() {
    print_status "Setting up secrets in Secret Manager..."
    
    # Create GitHub token secret
    if gcloud secrets describe github-token &>/dev/null; then
        print_success "GitHub token secret already exists"
    else
        print_status "Creating GitHub token secret..."
        echo -n "$GITHUB_TOKEN" | gcloud secrets create github-token --data-file=-
        print_success "GitHub token secret created"
    fi
    
    # Create webhook secret if provided
    if [[ -n "$GITHUB_WEBHOOK_SECRET" ]]; then
        if gcloud secrets describe github-webhook-secret &>/dev/null; then
            print_success "GitHub webhook secret already exists"
        else
            print_status "Creating GitHub webhook secret..."
            echo -n "$GITHUB_WEBHOOK_SECRET" | gcloud secrets create github-webhook-secret --data-file=-
            print_success "GitHub webhook secret created"
        fi
    fi
    
    # Grant Cloud Build access to secrets
    print_status "Granting Cloud Build access to secrets..."
    PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
    CLOUD_BUILD_SA="$PROJECT_NUMBER@cloudbuild.gserviceaccount.com"
    
    gcloud secrets add-iam-policy-binding github-token \
        --member="serviceAccount:$CLOUD_BUILD_SA" \
        --role="roles/secretmanager.secretAccessor" &>/dev/null || true
    
    if [[ -n "$GITHUB_WEBHOOK_SECRET" ]]; then
        gcloud secrets add-iam-policy-binding github-webhook-secret \
            --member="serviceAccount:$CLOUD_BUILD_SA" \
            --role="roles/secretmanager.secretAccessor" &>/dev/null || true
    fi
    
    # Grant Cloud Build additional roles
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$CLOUD_BUILD_SA" \
        --role="roles/run.admin" &>/dev/null || true
    
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$CLOUD_BUILD_SA" \
        --role="roles/iam.serviceAccountUser" &>/dev/null || true
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
    if [[ -f "package-lock.json" ]]; then
        npm ci
    else
        npm install
    fi

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

# Build and push Docker image
build_and_push_image() {
    print_status "Building and pushing Docker image..."
    
    # Generate image tags
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    IMAGE_BASE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${APP_NAME}"
    IMAGE_LATEST="${IMAGE_BASE}:latest"
    IMAGE_TAGGED="${IMAGE_BASE}:${TIMESTAMP}"

    # Build the image
    docker build -t "$IMAGE_LATEST" -t "$IMAGE_TAGGED" .

    # Push the images
    docker push "$IMAGE_LATEST"
    docker push "$IMAGE_TAGGED"

    print_success "Docker image built and pushed successfully"
    print_status "Image: $IMAGE_LATEST"
    print_status "Tagged: $IMAGE_TAGGED"
}

# Deploy to Cloud Run
deploy_to_cloud_run() {
    print_status "Deploying to Cloud Run..."
    
    IMAGE_URL="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${APP_NAME}:latest"
    
    # Prepare environment variables
    ENV_VARS="GITHUB_TOKEN=$$GITHUB_TOKEN,PORT=8080,GIN_MODE=release"
    if [[ -n "$GITHUB_WEBHOOK_SECRET" ]]; then
        ENV_VARS="$ENV_VARS,GITHUB_WEBHOOK_SECRET=$$GITHUB_WEBHOOK_SECRET"
    fi
    
    gcloud run deploy "$APP_NAME" \
        --image="$IMAGE_URL" \
        --region="$REGION" \
        --allow-unauthenticated \
        --set-env-vars="$ENV_VARS" \
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

# Set up Cloud Build trigger
setup_cloud_build_trigger() {
    print_status "Setting up Cloud Build trigger..."
    
    # Check if trigger already exists
    if gcloud builds triggers describe jgn-dev-deploy --region="$REGION" &>/dev/null; then
        print_success "Cloud Build trigger already exists"
        return
    fi
    
    print_warning "Cloud Build trigger setup requires manual configuration."
    print_status "Follow these steps in the GCP Console:"
    echo "1. Go to Cloud Build > Triggers"
    echo "2. Click 'Connect Repository' and connect your GitHub repository"
    echo "3. Create a trigger with these settings:"
    echo "   - Name: jgn-dev-deploy"
    echo "   - Event: Push to branch"
    echo "   - Branch: ^main$"
    echo "   - Configuration: cloudbuild.yaml"
    echo ""
    print_status "Or use the CI/CD guide: docs/cicd-guide.md"
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
    create_artifact_registry
    setup_secrets
    build_and_test
    build_and_push_image
    deploy_to_cloud_run
    setup_cloud_build_trigger

    echo -e "${GREEN}"
    echo "=================================================="
    echo "           Deployment Completed Successfully!    "
    echo "=================================================="
    echo -e "${NC}"
    
    print_status "Next steps:"
    echo "1. Set up Cloud Build trigger for automated deployments (see docs/cicd-guide.md)"
    echo "2. Configure GitHub webhook with the webhook URL shown above"
    echo "3. Update your domain DNS to point to the Cloud Run service"
    echo "4. Test the deployment by visiting your service URL"
    echo "5. Set up monitoring and alerts in GCP Console"
}

# Initialize infrastructure only
init_only() {
    print_status "Initializing GCP infrastructure only..."
    check_prerequisites
    validate_environment
    setup_gcp_auth
    create_artifact_registry
    setup_secrets
    setup_cloud_build_trigger
    print_success "Infrastructure initialization completed!"
}

# Handle script arguments
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  help    Show this help message"
        echo "  check   Check prerequisites only"
        echo "  init    Initialize infrastructure only (no deployment)"
        echo ""
        echo "Environment variables:"
        echo "  PROJECT_ID              GCP project ID (required)"
        echo "  GITHUB_TOKEN           GitHub personal access token (required)"
        echo "  GITHUB_WEBHOOK_SECRET  GitHub webhook secret (optional)"
        echo "  REGION                 GCP region (default: us-central1)"
        echo "  APP_NAME               Application name (default: jgn-dev)"
        echo "  REPO_NAME              Artifact Registry repo name (default: jgn-dev-repo)"
        ;;
    "check")
        check_prerequisites
        validate_environment
        ;;
    "init")
        init_only
        ;;
    *)
        main
        ;;
esac 