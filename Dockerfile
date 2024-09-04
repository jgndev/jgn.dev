# Amazon Linux 2 chosen for compatibility
FROM public.ecr.aws/amazonlinux/amazonlinux:2 AS builder

# Install Go
RUN yum update -y && \
    yum install -y golang && \
    yum clean all

# Set the working directory in the container to /app. \
# All following commands will be run from this directory.
WORKDIR /app

# Copy only the Go module fils first. This leverages Docker's cache layers
# to only re-fetch dependencies if these files change, optimizing build times.
COPY go.mod go.sum ./

# Download Go module dependencies. Separating this step to improve build times.
RUN go mod download

# Copy the application code into the container's /app directory. The build cache
# can be used to skip unchanged layers by running the copy as late as possible.
COPY . .

# Compile the application into a static binary for Cloud Run using Linux AMD64
# as the target build machine type.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./server/main.go

# Final state
FROM public.ecr.aws/amazonlinux/amazonlinux:2

# Install ca-certificates for HTTPS requests
RUN yum update -y && \
    yum install -y ca-certificates && \
    yum clean all

# Set the working directory inside the container to /app
# This affects subsequent RUN, CMD, ENTRYPOINT, COPY and ADD instructions
WORKDIR /app

# Copy the compiled binary 'main' from the builder stage to the current working directory (/app)
# This uses multi-stage builds to keep the final image small, containing only the necessary executable
COPY --from=builder /app/main .

# Copy over the static assets
COPY public/ /app/public/

# Cloud Run uses a PORT variablle which we can set explicitly here.
EXPOSE 8080

# Prevent the executable from being wrapped in a shell, reducing startup time
# and signal handling issues. Execute the compiled binary at /app/main.
CMD ["/app/main"]
