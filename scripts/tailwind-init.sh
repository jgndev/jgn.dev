#!/usr/bin/env bash

# Exit on error, undefined variable, or pipe failure
set -euo pipefail

# Script constants
SCRIPT_NAME=$(basename "$0")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
STYLE_FILE="$PROJECT_ROOT/public/css/style.css"
CONFIG_FILE="$PROJECT_ROOT/tailwind.config.js"

# Function to display usage information
usage() {
    echo "Usage: $SCRIPT_NAME"
    echo
    echo "Set up Tailwind CSS for the project."
    echo "This script should be run from the 'scripts' directory."
    echo "It installs Tailwind CSS, creates a configuration file,"
    echo "and sets up the necessary CSS file with Tailwind directives."
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

# Check if we're in the scripts directory
if [[ "$(basename "$PWD")" != "scripts" ]]; then
    error_exit "This script must be run from the 'scripts' directory."
fi

# Check for npm installation
if ! command -v npm &>/dev/null; then
    error_exit "npm could not be found. Please install Node.js and npm first."
fi

# Change to the project root directory
cd "$PROJECT_ROOT" || error_exit "Failed to change to project root directory."

# Install Tailwind CSS as a development dependency
log "Installing Tailwind CSS..."
npm install -D tailwindcss || error_exit "Failed to install Tailwind CSS."

# Initialize Tailwind CSS configuration
log "Initializing Tailwind CSS configuration..."
npx tailwindcss init || error_exit "Failed to initialize Tailwind CSS configuration."

# Generate the Tailwind CSS configuration with custom settings
log "Generating Tailwind CSS configuration file..."
cat > "$CONFIG_FILE" << EOF || error_exit "Failed to create $CONFIG_FILE"
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./internal/views/**/*.{html,js,templ}'],
    theme: {
        extend: {},
        fontFamily: {
            sans: ['JetBrains Mono', 'monospace'],
            serif: ['JetBrains Mono', 'monospace'],
            mono: ['JetBrains Mono', 'monospace'],
        },
    },
    plugins: [],
}
EOF

# Check if the style.css file already exists
if [ -f "$STYLE_FILE" ]; then
    log "$STYLE_FILE exists. Appending Tailwind directives..."
    # Check if Tailwind directives already exist in the file
    if grep -q "@tailwind" "$STYLE_FILE"; then
        log "Tailwind directives already present in $STYLE_FILE. Skipping append."
    else
        # Append Tailwind directives to the style.css file
        cat << EOF >> "$STYLE_FILE" || error_exit "Failed to append to $STYLE_FILE"

@tailwind base;
@tailwind components;
@tailwind utilities;
EOF
    fi
else
    log "Creating $STYLE_FILE with Tailwind directives..."
    # Create style.css file with Tailwind directives
    mkdir -p "$(dirname "$STYLE_FILE")" || error_exit "Failed to create directory for $STYLE_FILE"
    cat << EOF > "$STYLE_FILE" || error_exit "Failed to create $STYLE_FILE"
@tailwind base;
@tailwind components;
@tailwind utilities;
EOF
fi

log "✅ Tailwind CSS installation and configuration completed successfully."