# ===== Build Stage =====
FROM node:18-alpine AS css-builder

# Install Tailwind CSS
WORKDIR /app
COPY package*.json ./

# Use npm ci if lock file exists, otherwise use npm install
RUN if [ -f package-lock.json ]; then \
        npm ci --omit=dev; \
    else \
        npm install --omit=dev; \
    fi

# Copy CSS source and build CSS
COPY public/css/style.css ./public/css/
COPY internal/ ./internal/
RUN npx tailwindcss -i ./public/css/style.css -o ./public/css/site.css --minify

# ===== Go Build Stage =====
FROM golang:1.24-alpine AS go-builder

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate templ files
RUN templ generate

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o stratocraft-server ./server

# ===== Final Stage =====
FROM alpine:3.19

# Install only essential packages in single layer
RUN apk --no-cache add --update ca-certificates tzdata curl && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=go-builder /app/stratocraft-server .

# Copy only essential static assets
COPY --from=css-builder /app/public/css/site.css ./public/css/
COPY --from=go-builder /app/public/css/tokyo-night-dark.css ./public/css/
COPY --from=go-builder /app/public/js/ ./public/js/
COPY --from=go-builder /app/public/img/favicon.ico ./public/img/
COPY --from=go-builder /app/public/txt/robots.txt ./public/txt/

# Set proper permissions in single layer
RUN chown -R appuser:appgroup /app && \
    chmod +x ./stratocraft-server

# Switch to non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Optimized health check (less frequent = lower cost)
HEALTHCHECK --interval=60s --timeout=10s --start-period=60s --retries=2 \
    CMD curl -f http://localhost:8080/ || exit 1

# Environment variables for production
ENV GIN_MODE=release
ENV PORT=8080

# Run the application
CMD ["./stratocraft-server"]