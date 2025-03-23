#!/usr/bin/env bash

# Exit on error, undefined variable, or pipe failure
set -euo pipefail

# Script constants
SCRIPT_NAME=$(basename "$0")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
STYLE_FILE="$PROJECT_ROOT/public/css/style.css"
OUTPUT_FILE="$PROJECT_ROOT/public/css/site.css"

# Function to display usage information
usage() {
    echo "Usage: $SCRIPT_NAME"
    echo
    echo "Watch for changes in style.css and compile Tailwind CSS."
    echo "This script should be run from the 'scripts' directory."
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

# Function to check if a command/software is installed
is_installed() {
    command -v "$1" &>/dev/null
}

# Check for npx installation
if ! is_installed npx; then
    error_exit "npx is not installed. Please install Node.js which provides npx."
fi

# Check if the style file exists
if [ ! -f "$STYLE_FILE" ]; then
    error_exit "$STYLE_FILE does not exist. Please check the file path or create it."
fi

# Check if Tailwind CSS is installed
if ! npx tailwindcss --help &>/dev/null; then
    error_exit "Tailwind CSS is not installed. Please run 'npm install tailwindcss' first."
fi

# Ensure output directory exists
mkdir -p "$(dirname "$OUTPUT_FILE")" || error_exit "Failed to create output directory."

# Run Tailwind CSS in watch mode with minification
log "Watching for changes in $STYLE_FILE. Press CTRL+C to stop."
npx tailwindcss -i "$STYLE_FILE" -o "$OUTPUT_FILE" --minify --watch