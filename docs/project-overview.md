# jgn.dev Project Overview

This document provides a comprehensive overview of the jgn.dev project - a modern, high-performance blog and cheat sheet site built with Go, Templ templates, and Tailwind CSS.

## 🎯 Project Vision

Create a lightning-fast, minimalist blog and cheat sheet platform that achieves perfect Lighthouse scores while maintaining developer productivity and content management simplicity.

## 🏗️ Architecture Overview

### High-Level Architecture

```
┌────────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ GitHub Posts &     │    │     jgn.dev      │    │   GitHub        │
│ Cheat Sheets Repos │───▶│   Application    │◀───│   Webhook       │
│ (Markdown)         │    │   (Go + Templ)   │    │   (Auto-refresh)│
└────────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Content API   │    │   GCP Cloud Run  │    │   Real-time     │
│   (GitHub API)  │    │   (Containers)   │    │   Search        │
│                 │    │                  │    │   (HTMX)        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### Technology Stack

**Backend Technologies:**
- **Go 1.24**: Core application language
- **Echo v4**: High-performance HTTP framework
- **Templ**: Type-safe HTML template engine
- **GitHub API**: Content management and retrieval (posts and cheat sheets)

**Frontend Technologies:**
- **Tailwind CSS v4**: Utility-first CSS framework
- **HTMX**: Dynamic HTML interactions
- **Highlight.js**: Syntax highlighting for code blocks
- **Vanilla JavaScript**: Theme switching and mobile navigation

**Infrastructure:**
- **Docker**: Containerization with multi-stage builds
- **GCP Cloud Run**: Serverless container hosting
- **GCP Artifact Registry**: Container image storage
- **GCP Cloud Build**: Automated CI/CD pipeline

## 🚀 Key Features

### Content Management
- **GitHub-Based CMS**: Posts and cheat sheets stored as Markdown files in separate repositories
- **Webhook Integration**: Automatic content refresh on GitHub pushes
- **Frontmatter Support**: YAML metadata for posts and cheat sheets (title, date, author, tags, summary, etc.)
- **Real-time Updates**: No server restarts required for content changes

### User Experience
- **Lightning Fast**: Optimized for 100 Lighthouse Performance score
- **Real-time Search**: HTMX-powered search with instant results for both posts and cheat sheets
- **Cheat Sheets Section**: Browse and search quick reference guides at `/cheatsheets`
- **Mobile-First Design**: Responsive layout with mobile navigation
- **Dark/Light Themes**: Enhanced dark mode with deeper colors
- **Accessibility**: Full ARIA compliance and keyboard navigation

### Developer Experience
- **Type-Safe Templates**: Templ provides compile-time template safety
- **Hot Reloading**: Automatic recompilation during development
- **Comprehensive Tooling**: Linting, formatting, and testing scripts
- **Docker Integration**: Consistent development and production environments

### Performance Features
- **Optimized Docker Images**: Multi-stage builds with minimal runtime footprint
- **Efficient Asset Loading**: Font subsetting and CSS minification
- **Consistent Syntax Highlighting**: Tokyo Night Dark theme across light/dark modes
- **Gzip Compression**: Built-in response compression

## 📁 Project Structure Deep Dive

### Core Application (`internal/`)

```
internal/
├── application/           # HTTP handlers and controllers
│   ├── home.go           # Home page with recent posts
│   ├── posts.go          # Posts listing with pagination
│   ├── post.go           # Individual post rendering
│   ├── search.go         # Real-time search functionality
│   ├── about.go          # About page handler
│   └── webhook.go        # GitHub webhook handler
├── contentmanager/       # Content fetching and parsing
│   ├── contentmanager.go # GitHub API integration
│   └── parsemarkdown.go  # Markdown processing with frontmatter
├── views/                # Templ templates
│   ├── pages/            # Full page templates
│   ├── components/       # Reusable UI components
│   └── shared/           # Layout and navigation
└── site/                 # Configuration and metadata
```

### Static Assets (`public/`)

```
public/
├── css/
│   ├── style.css         # Source Tailwind CSS
│   ├── site.css          # Compiled/minified CSS
│   └── tokyo-night-dark.css # Syntax highlighting theme
├── js/
│   ├── theme.js          # Dark/light mode switching
│   └── mobile-nav.js     # Mobile navigation behavior
├── font/
│   ├── Inter-*.woff2     # Inter font family
│   └── jetbrains-mono-*.woff2 # JetBrains Mono for code
├── img/
│   └── favicon.ico       # Site favicon
└── txt/
    └── robots.txt        # Search engine directives
```

### Development Tools (`scripts/`)

```
scripts/
├── run-dev.sh           # All-in-one development environment
├── deploy-gcp-cloud-run.sh # GCP deployment automation
├── templ-watch.sh       # Template compilation watching
├── tailwind-watch.sh    # CSS compilation watching
└── test-webhook.sh      # Webhook testing utility
```

## 🔧 Configuration Management

### Environment Variables

**Required:**
- `GITHUB_TOKEN`: GitHub Personal Access Token for API access
- Project requires minimal configuration for maximum simplicity

**Optional:**
- `GITHUB_WEBHOOK_SECRET`: Secret for webhook signature verification
- `PORT`: Server port (defaults to 8080)

### Site Configuration (`internal/site/site.go`)

Centralized configuration for:
- GitHub repository settings
- Site metadata and branding
- Navigation structure
- Social media links

## 🚀 Deployment Architecture

### GCP Cloud Run Setup

**Service Configuration:**
- **CPU**: 1 vCPU (scalable based on traffic)
- **Memory**: 512Mi (optimized for Go application)
- **Scaling**: 0-10 instances (cost-effective auto-scaling)
- **Region**: us-central1 (configurable)

**Container Optimization:**
- Multi-stage Docker builds
- Alpine Linux base (minimal attack surface)
- Non-root user execution
- Health check endpoints

### CI/CD Pipeline

**Cloud Build Triggers:**
- Automatic builds on main branch pushes
- Docker image building and registry push
- Zero-downtime Cloud Run deployment
- Post-deployment health verification

**Build Process:**
1. Node.js stage: Tailwind CSS compilation
2. Go stage: Templ generation and binary compilation
3. Final stage: Minimal runtime image assembly
4. Deployment: Cloud Run service update

## 📊 Performance Metrics

### Lighthouse Scores (GCP Cloud Run)
- **Performance**: 100/100
- **Accessibility**: 100/100
- **Best Practices**: 96/100
- **SEO**: 91/100

### Load Performance
- **First Contentful Paint**: < 1.5s
- **Largest Contentful Paint**: < 2.5s
- **Time to Interactive**: < 3.0s
- **Total Blocking Time**: < 200ms

### Resource Optimization
- **Docker Image Size**: ~50MB (multi-stage optimization)
- **CSS Bundle Size**: ~10KB (Tailwind purging)
- **JavaScript Bundle**: ~5KB (minimal client-side code)
- **Font Loading**: Optimized WOFF2 with font-display: swap

## 🔐 Security Architecture

### Application Security
- **Non-root Container Execution**: Security best practice
- **Minimal Attack Surface**: Alpine Linux base image
- **Input Validation**: GitHub webhook signature verification
- **HTTPS Only**: Cloud Run enforces TLS

### Access Control
- **GitHub API**: Token-based authentication
- **Webhook Security**: HMAC signature verification
- **Cloud Run**: Public read access, protected admin endpoints

## 🎨 Design System

### Typography
- **Primary Font**: Inter (web-optimized variable font)
- **Code Font**: JetBrains Mono (programming-focused)
- **Font Loading**: Optimized with font-display: swap

### Color System
- **Light Mode**: Clean whites and subtle grays
- **Dark Mode**: Deep blacks (#0a0a0a) with enhanced contrast
- **Accent Colors**: Consistent brand colors across themes

### Component Architecture
- **Responsive Design**: Mobile-first approach
- **Component Reusability**: Shared Templ components
- **Accessibility**: Full ARIA compliance

## 💰 Cost Structure

### GCP Cloud Run Pricing (Estimated)
- **Low Traffic** (1K visits/month): $0-2/month
- **Medium Traffic** (10K visits/month): $2-8/month
- **High Traffic** (100K visits/month): $15-30/month

### Cost Optimization Features
- **Scale to Zero**: No charges when inactive
- **Efficient Resource Usage**: Optimized memory and CPU usage
- **Minimal Network Egress**: Optimized asset delivery

## 🔄 Content Workflow

### Publishing Process
1. **Content Creation**: Write Markdown posts in the posts repo, or cheatsheets in the cheatsheets repo ([github.com/jgndev/cheatsheets](https://github.com/jgndev/cheatsheets))
2. **Automatic Processing**: Webhook triggers content refresh for both posts and cheat sheets
3. **Live Updates**: Site updates without manual deployment
4. **Search Indexing**: Real-time search index updates for both posts and cheat sheets

### Content Features
- **Frontmatter Parsing**: YAML metadata extraction for both posts and cheat sheets
- **Tag System**: Categorization and filtering for both content types
- **Date Handling**: RFC3339 timestamp parsing
- **Slug Generation**: URL-friendly post and cheat sheet identifiers

### Cheat Sheet Example Frontmatter
```yaml
---
id: unique-cheatsheet-id
title: "Go Slices Cheat Sheet"
date: 2024-01-20T00:00:00Z
author: "Jeremy Novak"
tags: ["go", "slices", "reference"]
slug: go-slices
published: true
summary: "Quick reference for Go slices syntax and operations."
---
```

## 🚨 Monitoring and Maintenance

### Health Monitoring
- **Container Health Checks**: HTTP endpoint monitoring
- **Cloud Run Metrics**: Request latency and error rates
- **Build Pipeline Monitoring**: Cloud Build success rates

### Maintenance Tasks
- **GitHub Token Rotation**: Periodic token updates
- **Container Image Updates**: Base image security updates
- **Performance Monitoring**: Lighthouse score tracking

## 🎯 Future Roadmap

### Planned Enhancements
- **RSS Feed**: XML feed generation for syndication
- **Comment System**: Integration with external comment services
- **Analytics**: Privacy-focused analytics integration
- **CDN Integration**: Global content delivery optimization

### Technical Improvements
- **Advanced Caching**: Redis integration for content caching
- **Image Optimization**: Automatic image resizing and WebP conversion
- **Progressive Web App**: Service worker and offline capabilities
- **Advanced Search**: Full-text search with fuzzy matching

---

This overview provides a comprehensive understanding of the jgn.dev project's current state, architecture, and capabilities. For specific implementation details, refer to the individual documentation files in the `docs/` directory. 