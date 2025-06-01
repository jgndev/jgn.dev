# jgn.dev

A modern, high-performance blog built with Go, Templ templates, and Tailwind CSS. Features real-time search, GitHub-based content management, and automatic post updates via webhooks.

You can give the live site a spin at <a href="https://jgn.dev" target="_blank">jgn.dev</a>

## ⚡ Performance First

Designed and optimized for exceptional web performance:

🏆 **100 Lighthouse Score** for Performance when deployed to GCP Cloud Run.

![Lighthouse Score](docs/gcp-cloud-run-lighthouse-score.png)

## 🚀 Features

- **Modern Go Stack**: Go 1.24 + Echo v4 + Templ templates + Tailwind CSS v4
- **GitHub Content Management**: Posts stored as Markdown in GitHub repository
- **Real-time Search**: HTMX-powered search with tag and title filtering
- **Webhook Auto-Updates**: Automatically refreshes content when posts are added to GitHub
- **Responsive Design**: Mobile-first design with dark/light mode support
- **Syntax Highlighting**: Code blocks with Tokyo Night Dark theme (consistent across light/dark modes)
- **SEO Optimized**: Structured data, meta tags, and semantic HTML
- **Production Ready**: Docker containerization for GCP Cloud Run deployment

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   GitHub Posts  │    │     jgn.dev      │    │   GitHub        │
│   Repository    │───▶│   Application    │◀───│   Webhook       │
│   (Markdown)    │    │   (Go + Templ)   │    │   (Auto-refresh)│
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Content API   │    │   GCP Cloud Run  │    │   Real-time     │
│   (GitHub API)  │    │   (Containers)   │    │   Search        │
│                 │    │   + Terraform    │    │   (HTMX)        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🛠️ Technology Stack

**Backend:**
- Go 1.24 with Echo v4 framework
- Templ for type-safe HTML templates
- GitHub API for content management
- HTMX for dynamic interactions

**Frontend:**
- Tailwind CSS v4 for styling
- Vanilla JavaScript for theme switching
- Highlight.js for syntax highlighting
- Responsive design with mobile navigation

**Infrastructure:**
- Docker for containerization
- GCP Cloud Run for hosting
- GCP Artifact Registry for image storage
- Terraform for infrastructure as code
- GitHub Actions for CI/CD pipeline

## 📁 Project Structure

```
jgn.dev/
├── docs/
│   ├── gcp-deployment-guide.md              # GCP Cloud Run deployment guide
│   ├── terraform-guide.md                   # Terraform infrastructure guide
│   ├── cicd-guide.md                        # GitHub Actions CI/CD guide
│   ├── webhook-setup-guide.md               # GitHub webhook setup
│   └── github-token-setup-guide.md          # GitHub token configuration
├── terraform/
│   ├── main.tf                              # Main Terraform configuration
│   ├── variables.tf                         # Terraform variables
│   ├── outputs.tf                           # Terraform outputs
│   └── terraform.tfvars.example             # Example configuration
├── .github/workflows/
│   └── deploy.yml                           # GitHub Actions CI/CD pipeline
├── internal/
│   ├── application/                         # HTTP handlers
│   │   ├── home.go                          # Home page handler
│   │   ├── posts.go                         # Posts listing handler
│   │   ├── post.go                          # Individual post handler
│   │   ├── search.go                        # Search functionality
│   │   ├── about.go                         # About page handler
│   │   └── webhook.go                        # GitHub webhook handler
│   ├── contentmanager/                       # GitHub integration
│   │   ├── contentmanager.go                 # Content fetching logic
│   │   └── parsemarkdown.go                  # Markdown parsing
│   ├── views/
│   │   ├── pages/                             # Page templates
│   │   ├── components/                          # Reusable components
│   │   └── shared/                              # Layout and navigation
│   └── site/                                    # Site configuration
├── public/
│   ├── css/                                      # Stylesheets
│   ├── js/                                         # JavaScript files
│   └── img/                                        # Images and assets
├── scripts/
│   ├── deploy-gcp-cloud-run.sh                      # GCP Cloud Run deployment
│   ├── run-dev.sh                                     # Development orchestration
│   ├── templ-watch.sh                                 # Template hot reloading
│   ├── tailwind-watch.sh                                # CSS hot reloading
│   └── test-webhook.sh                                    # Webhook testing
├── server/
│   └── main.go                                         # Application entry point
├── Dockerfile                                            # Container configuration
├── package.json                                            # Tailwind CSS dependencies
└── README.md                                                 # This file
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker (for deployment)
- Node.js (for Tailwind CSS)
- gcloud CLI (for GCP deployment)
- Terraform (for infrastructure)
- GitHub Personal Access Token ([setup guide](docs/github-token-setup-guide.md))

### Local Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/jgndev/jgn.dev.git
   cd jgn.dev
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go install github.com/a-h/templ/cmd/templ@latest
   go install github.com/cosmtrek/air@latest
   npm install
   ```

3. **Set environment variables**
   ```bash
   export GITHUB_TOKEN=your_github_token_here  # See: docs/github-token-setup-guide.md
   export GITHUB_WEBHOOK_SECRET=$(openssl rand -hex 32)  # Optional
   ```

4. **Start development environment**
   
   **Option A: All-in-One Script (Recommended)**
   ```bash
   ./scripts/run-dev.sh
   ```
   
   This single command:
   - ✅ Checks prerequisites and environment
   - ✅ Runs initial builds
   - ✅ Starts Tailwind CSS watch (background)
   - ✅ Starts Templ template watch (background)  
   - ✅ Starts Air Go hot reload (foreground)
   - ✅ Provides proper cleanup on Ctrl+C
   
   **Additional commands:**
   ```bash
   ./scripts/run-dev.sh stop      # Stop all processes
   ./scripts/run-dev.sh restart   # Restart all processes
   ./scripts/run-dev.sh status    # Show process status
   ```
   
   **Option B: Manual Setup (Advanced)**
   
   If you prefer to run each process manually in separate terminals:
   
   **Terminal 1 - Templ Watch:**
   ```bash
   ./scripts/templ-watch.sh
   ```
   
   **Terminal 2 - Tailwind Watch:**
   ```bash
   ./scripts/tailwind-watch.sh
   ```
   
   **Terminal 3 - Go Server with Air:**
   ```bash
   air
   ```

5. **Visit** http://localhost:8080

   The application will automatically reload when you make changes to:
   - Go files (via Air)
   - Templ templates (via templ-watch.sh)
   - CSS styles (via tailwind-watch.sh)

## 🌐 Deployment to GCP Cloud Run

Deploy to GCP Cloud Run for cost-effective, scalable hosting with automatic CI/CD:

### Quick Deployment
```bash
export PROJECT_ID=your-gcp-project-id
export GITHUB_TOKEN=your_token
export GITHUB_WEBHOOK_SECRET=your_webhook_secret
./scripts/deploy-gcp-cloud-run.sh
```

**Why Cloud Run:**
- ✅ Pay-per-use pricing (can cost as little as $0-5/month for low traffic)
- ✅ Automatic scaling to zero when not in use
- ✅ Built-in SSL certificates and global load balancing
- ✅ Easy custom domain configuration
- ✅ Integrated CI/CD with Cloud Build
- ✅ Better cost efficiency compared to always-on services
- ✅ Built-in health monitoring and auto-restart

📖 **Detailed Instructions**: See [GCP Deployment Guide](docs/gcp-deployment-guide.md) for complete setup instructions and configuration options.

## 🔄 GitHub Actions CI/CD

Automatic deployment pipeline that triggers on every push to main:

### Features
- ✅ **Automated Testing**: Go tests, linting, and format checking
- ✅ **Security Scanning**: Trivy vulnerability scanning on PRs
- ✅ **Docker Build**: Multi-stage builds with caching
- ✅ **Cloud Run Deployment**: Automatic deployment to GCP
- ✅ **Smoke Tests**: Post-deployment health checks
- ✅ **Status Updates**: Commit status updates with deployment results

### Required Secrets
Configure these secrets in your GitHub repository:
- `GCP_PROJECT_ID`: Your GCP project ID
- `GCP_REGION`: GCP region (e.g., us-central1)
- `GCP_SA_KEY`: GCP service account key (JSON)
- `GITHUB_TOKEN_FOR_API`: GitHub token for API access
- `GITHUB_WEBHOOK_SECRET`: Webhook secret for GitHub integration

📖 **Detailed Instructions**: See [CI/CD Guide](docs/cicd-guide.md) for complete pipeline setup and configuration.

## 📊 Infrastructure as Code

Complete Terraform configuration for reproducible deployments:

### Resources Created
- **Cloud Run Service**: Serverless container hosting
- **Artifact Registry**: Private container image repository
- **Service Accounts**: Secure identity and access management
- **Cloud Build Triggers**: Automated CI/CD integration
- **Domain Mapping**: Custom domain configuration (optional)

### Deploy Infrastructure
```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your configuration
terraform init
terraform plan
terraform apply
```

📖 **Detailed Instructions**: See [Terraform Guide](docs/terraform-guide.md) for infrastructure setup and management.

## 🔄 GitHub Webhook Setup

Enable automatic content updates when you add new blog posts:

1. Generate a webhook secret: `openssl rand -hex 32`
2. Configure the webhook in your GitHub posts repository
3. Set the webhook URL to: `https://yourdomain.com/webhook/github`

📖 **Detailed Instructions**: See [Webhook Setup Guide](docs/webhook-setup-guide.md) for step-by-step configuration.

## ✨ Key Features

### Content Management
- **GitHub Integration**: Posts stored as Markdown files in a separate GitHub repository
- **Automatic Refresh**: Webhook-triggered content updates without server restarts
- **Frontmatter Support**: YAML frontmatter for post metadata (title, date, tags, etc.)

### Search & Navigation
- **Real-time Search**: HTMX-powered search across post titles and tags
- **Posts Listing**: Paginated view of all posts, sorted by date
- **Individual Post Pages**: Clean, readable post layout with syntax highlighting
- **Mobile Navigation**: Hamburger menu with smooth animations

### Performance & SEO
- **Fast Loading**: Optimized Docker images and efficient Go backend
- **Consistent Syntax Highlighting**: Tokyo Night Dark theme in both light and dark modes
- **Dark/Light Mode**: Enhanced dark mode with deeper colors
- **Responsive Design**: Mobile-first approach with Tailwind CSS

## 🔧 Configuration

### Environment Variables

**Required:**
- `GITHUB_TOKEN`: Personal access token for GitHub API access ([setup guide](docs/github-token-setup-guide.md))

**Optional:**
- `GITHUB_WEBHOOK_SECRET`: Secret for webhook signature verification
- `PORT`: Server port (default: 8080)

### Site Configuration

Edit `internal/site/site.go` to configure:
- Post repository owner and name
- Site metadata and branding
- Navigation links

## 📝 Content Management

### Adding Blog Posts

1. **Create a Markdown file** in your posts repository
2. **Add frontmatter**:
   ```yaml
   ---
   id: unique-post-id
   title: "Your Post Title"
   date: 2024-01-15T00:00:00Z
   tags: ["go", "web-development", "gcp"]
   slug: your-post-slug
   published: true
   ---
   ```
3. **Write your content** in Markdown
4. **Commit and push** - the site will automatically update via webhook  

💡 Check out the posts repository at [github.com/jgndev/posts](https://github.com/jgndev/posts) for examples.

### Supported Frontmatter Fields

- `id`: Unique identifier for the post
- `title`: Post title (required)
- `date`: Publication date in RFC3339 format
- `tags`: Array of tags for categorization
- `slug`: URL slug (auto-generated if not provided)
- `published`: Boolean to control post visibility

## 🐳 Docker

### Build Image
```bash
docker build -t jgn-dev .
```

### Run Container
```bash
docker run -p 8080:8080 \
  -e GITHUB_TOKEN=your_token \
  -e GITHUB_WEBHOOK_SECRET=your_secret \
  jgn-dev
```

## 🧪 Testing

### Test Webhook Locally
```bash
export GITHUB_WEBHOOK_SECRET=your_secret
./scripts/test-webhook.sh
```

### Run Application Tests
```bash
go test ./...
```

## 🚨 Troubleshooting

### Common Issues

1. **Templ generation fails**
   - Install templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`
   - Run `templ generate` before building

2. **Tailwind styles not loading**
   - Build CSS: `npx tailwindcss -i ./public/css/style.css -o ./public/css/site.css --minify`
   - Check file paths in templates

3. **GitHub API rate limiting**
   - Set `GITHUB_TOKEN` environment variable ([setup guide](docs/github-token-setup-guide.md))
   - Verify token has repository read permissions

4. **Webhook not working**
   - Check webhook secret matches environment variable
   - Verify webhook URL is publicly accessible
   - Check server logs for error messages

5. **GCP deployment issues**
   - Verify gcloud authentication: `gcloud auth list`
   - Check project permissions and enabled APIs
   - Review Cloud Run service logs in GCP Console

## 💰 Cost Optimization

GCP Cloud Run pricing is usage-based, making it very cost-effective:

- **CPU**: $0.00002400 per vCPU-second
- **Memory**: $0.00000250 per GiB-second
- **Requests**: $0.40 per million requests
- **Networking**: Egress charges apply

**Estimated costs for a personal blog:**
- **Low traffic** (1K visits/month): $0-2/month
- **Medium traffic** (10K visits/month): $2-8/month
- **High traffic** (100K visits/month): $15-30/month

## 📚 Documentation

- **[GCP Deployment Guide](docs/gcp-deployment-guide.md)**: Complete GCP Cloud Run deployment instructions
- **[Terraform Guide](docs/terraform-guide.md)**: Infrastructure as code setup and management
- **[CI/CD Guide](docs/cicd-guide.md)**: GitHub Actions pipeline configuration
- **[GitHub Token Setup Guide](docs/github-token-setup-guide.md)**: Step-by-step guide to create and configure GitHub Personal Access Tokens
- **[Webhook Setup Guide](docs/webhook-setup-guide.md)**: GitHub webhook configuration for automatic updates

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.

## 📞 Support

- **Documentation**: Check the guides in the `docs/` folder
- **Issues**: Open a GitHub issue for bugs or feature requests
- **Contact**: [Open an issue](https://github.com/jgndev/jgn.dev/issues) for support

---

Built with 🤓 using Go, Templ, Tailwind CSS, and deployed on GCP Cloud Run