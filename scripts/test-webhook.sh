#!/bin/bash

# Test script for GitHub webhook endpoint
# This simulates what GitHub would send when a markdown file is added

if [ -z "$GITHUB_WEBHOOK_SECRET" ]; then
    echo "Error: GITHUB_WEBHOOK_SECRET environment variable must be set"
    echo "Generate one with: openssl rand -hex 32"
    exit 1
fi

# Sample webhook payload (what GitHub would send)
PAYLOAD='{
  "ref": "refs/heads/main",
  "commits": [
    {
      "added": ["new-test-post.md"],
      "modified": [],
      "removed": []
    }
  ],
  "repository": {
    "name": "posts",
    "full_name": "stratocraft/posts"
  }
}'

# Calculate HMAC signature
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$GITHUB_WEBHOOK_SECRET" | cut -d' ' -f2)

# Send the webhook request
echo "Testing webhook endpoint..."
echo "Payload: $PAYLOAD"
echo ""

curl -X POST http://localhost:8080/webhook/github \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n"

echo ""
echo "Check your server logs to see if the webhook was processed correctly." 