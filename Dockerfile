# Final image size will be approximately 20MB.

# Stage 1: Build Stage
# Using golang:1-23-alpine as ou builder image
# Alpine is chosen for its small size (~5MB) while still providing
# a full Linux environment. The golang image includes all necessary
# Go tools and dependencies.
FROM golang:1.23-alpine AS builder

# Build image dependencies:
# Install necessary build dependencies.
# ca-certificates: Needed for HTTPS requests during go mod download.
RUN apk add --no-cache ca-certificates

# WORKDIR:
# Set the working directory inside the container.
# This affects all subsequent commands in this build stage.
WORKDIR /app

# Copy dependencies:
# This is a Docker best practice that takes advantage of layer caching.
# If these files don't change, Docker will used the cached layer for
# dependency downloads.
COPY go.mod go.sum ./

# Download go mod dependencies:
# This step is separate from the build to take advantage of Docker's
# layer caching. Dependencies will only be re-downloaded if go.mod
# or go.sum change.
RUN go mod download

# Copy the source code:
# This happens after dependency download because source code changes
# more frequently and this keeps builds faster by not invalidating
# the dependency layer cache.
COPY . .

# Build the application:
# Build flags:
# CGO_ENABLED=0: Disable CGO, creating a statically linked binary.
# GOOS=linux: Target Linux OS
# GOARCH=amd64: Target amd64 architecture
# -ldflags='-w -s': Strip debug information to reduce binary size.
# -w: Disable DWARF generation
# -s: Disable symbol table
# -extldflags "-static": Use static linking
# -a: Force rebuild of packages that are already up-to-date.
# -installsuffix cgo: Add suffix to package directory to avoid conflicts.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./server/main.go

# Stage 2: Final Stage
# Using Google's distroless image as our final base.
# Distroless images contain only your application and its runtime dependencies.
# No package manager, shell, or other programs typically found on Linux.
FROM gcr.io/distroless/static:nonroot

# Copy only the built executable from builder stage:
# This is the power of multi-stage builds - we only take what we need.
# The final image won't contain any of the build tools or source code.
COPY --from=builder /app/main /app/main

# Copy static assets.
COPY public /app/public/

# Switch to non-root user for security:
# The nonroot user is provided by the distroless image.
# Running containers as non-root is a security best practice.
USER nonroot:nonroot

# Set the working directory for the application
WORKDIR /app

# Document that the container listens on port 8080:
# This is purely informational and doesn't actually expose the port.
EXPOSE 8080

# Define the command to run your application:
# Using [] syntax instead of shell form for better signal handling.
# This runs the binary directly, not wrapped in a shell.
CMD ["/app/main"]
