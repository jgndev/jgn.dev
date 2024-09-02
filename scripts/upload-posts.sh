#!/usr/bin/env bash

set -euo pipefail

# Script constants
SCRIPT_NAME=$(basename "$0")
ENV_FILE=".env"

# Function to display usage information
usage() {
    echo "Usage: $SCRIPT_NAME <path-to-markdown-file> <s3-key>"
    echo
    echo "Upload a Markdown file to an S3 bucket."
    echo
    echo "Arguments:"
    echo "  <path-to-markdown-file>  Path to the Markdown file to upload"
    echo "  <s3-key>                 S3 object key (path in the bucket) for the uploaded file"
    echo
    echo "Environment:"
    echo "  Requires S3_BUCKET_NAME to be set in $ENV_FILE file"
}

# Function to log messages
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" >&2
}

# Function to handle errors
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Check for correct number of arguments
if [ "$#" -ne 2 ]; then
    usage
    exit 1
fi

# Assign arguments to named variables
file_path="$1"
s3_key="$2"

# Check if file exists
[ -f "$file_path" ] || error_exit "File not found: $file_path"

# Check if .env file exists
[ -f "$ENV_FILE" ] || error_exit "$ENV_FILE file not found"

# Extract S3 bucket name from .env file
bucket_name=$(grep -E '^S3_BUCKET_NAME=' "$ENV_FILE" | cut -d '=' -f2)
[ -n "$bucket_name" ] || error_exit "S3_BUCKET_NAME not found in $ENV_FILE"

# Remove any surrounding quotes from bucket name
bucket_name=$(echo "$bucket_name" | sed -e 's/^"//' -e 's/"$//')

log "Uploading $file_path to s3://$bucket_name/$s3_key"

# Attempt to upload file to S3
if aws s3 cp "$file_path" "s3://$bucket_name/$s3_key"; then
    log "Upload successful"
else
    error_exit "Failed to upload file to S3"
fi