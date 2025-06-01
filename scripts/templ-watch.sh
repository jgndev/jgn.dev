#!/usr/bin/env bash

# Exit on any failures
set -e

# Function to check if a command or software is installed
is_installed() {
  command -v "$1" &>/dev/null
}

# Check for templ installation
if ! is_installed templ; then
  echo "The templ command could not be found. If you previously installed temp please check GOPATH."
  echo "If you have not yet installed templ, visit https://github.com/a-h/templ for instructions."
  exit 1
fi

# Start the watch process to build templ files on change
echo "Watching for template changes. Press CTRL+C to stop."
templ fmt internal/views/** && templ generate --watch
