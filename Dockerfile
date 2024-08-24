# Alpine Linux chosen for small base image size and fast startup times
FROM golang:1.21-alpine AS builder

# Set the working directory in the container to /app. All following commands
# will be run from this directory.
WORKDIR /app

# Copy only the Go module fils first. This leverages Docker's cache layers
# to only re-fetch dependencies if these files change, optimizing build times.
COPY go.mod ./
COPY go.sum ./

# Download Go module dependencies. Separating this step to improve build times.
RUN go mod download

# Copy the application code into the container's /app directory. The build cache
# can be used to skip unchanged layers by running the copy as late as possible.
COPY . ./

# Compile the application into a static binary for Cloud Run using Linux AMD64
# as the target build machine type.
# RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o server ./server/main.go
RUN go build -v ./server/main.go

# Cloud Run uses a PORT variablle which we can set explicitly here.
ENV PORT 8080
EXPOSE 8080

# Prevent the executable from being wrapped in a shell, reducing startup time
# and signal handling issues. Execute the compiled binary at /app/server.
CMD ["/app/main"]
