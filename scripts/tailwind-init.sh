#!/usr/bin/env bash

# Exit on error
set -e

# Verify npm installation
if ! command -v npm &>/dev/null; then
  echo "npm could not be found. Please install Node.js and npm first"
  exit 1
fi

# Install Tailwind CSS as a development dependency
echo "Installing Tailwind CSS..."
npm install -D tailwindcss

# Initialize Tailwind CSS configuration
echo "Initiliazing Tailwind CSS configuration..."
npx tailwindcss init

# Generate the Tailwind CSS configuration with custom settings
echo "Generating Tailwind CSS configuration file..."
echo <<EOF >tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./internal/views/**/*.{html,js,templ}'],
    theme: {
        extend: {},
        fontFamily: {
            'sans': ['Bahnschrift', 'DIN Alternate', 'Franklin Gothic Medium', 'Nimbus Sans Narrow', 'sans-serif-condensed', 'sans-serif'],
        },
    },
    plugins: [],
}
EOF

# Check if the style.css file aleady exists to prevent overwriting
STYLE_FILE="public/css/style.css"
if [ -f "$STYLE_FILE" ]; then
  echo "$STYLE_FILE exists. Appending Tailwind directives..."
else
  echo "Creating $STYLE_FILE with Tailwind directives..."
fi

# Append Tailwind directives to the style.css file
cat <<EOF >>$STYLE_FILE

@tailwind base;
@tailwind components;
@tailwind utilities;
EOF

echo "✅ Tailwind CSS installation and configuration copmleted."
