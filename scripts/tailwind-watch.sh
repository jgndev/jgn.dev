#!/usr/bin/env bash

# Script to watch style changes in public/css/style.css

# Exit on error
set -e

# Function to check if a command/software is installed
is_installed() {
  command -v "$1" &>/dev/null
}

# Check for npx installation
if ! is_installed npx; then
  echo "npx is not installed. Please install Node.js which provides npx."
  exit 1
fi

# Check if the style file exists
STYLE_FILE="./public/css/style.css"
if [ ! -f "$STYLE_FILE" ]; then
  echo "$STYLE_FILE does not exist. Please check the file path or create ./public/css/style.css."
  exit 1
fi

# Run Tailwind CSS in watch mode with minification
echo "Watching for changes in $STYLE_FILE. Press CTRL+C to stop."
npx tailwindcss -i $STYLE_FILE -o ./public/css/site.css --minify --watch
