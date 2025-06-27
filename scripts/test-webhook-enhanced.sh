#!/bin/bash

# Enhanced test script for GitHub webhook endpoint
# Tests both posts and cheatsheets repositories

if [ -z "$GITHUB_WEBHOOK_SECRET" ]; then
    echo "Error: GITHUB_WEBHOOK_SECRET environment variable must be set"
    echo "Generate one with: openssl rand -hex 32"
    exit 1
fi

WEBHOOK_URL="http://localhost:8080/webhook/github"

# Test function
test_webhook() {
    local repo_name=$1
    local repo_full_name=$2
    local test_file=$3
    
    echo "========================================="
    echo "Testing webhook for $repo_name repository"
    echo "========================================="
    
    # Sample webhook payload
    PAYLOAD='{
  "ref": "refs/heads/main",
  "commits": [
    {
      "added": ["'$test_file'"],
      "modified": [],
      "removed": []
    }
  ],
  "repository": {
    "name": "'$repo_name'",
    "full_name": "'$repo_full_name'"
  }
}'

    # Calculate HMAC signature
    SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$GITHUB_WEBHOOK_SECRET" | cut -d' ' -f2)

    echo "Repository: $repo_name"
    echo "Test file: $test_file"
    echo ""

    # Send the webhook request
    RESPONSE=$(curl -s -X POST "$WEBHOOK_URL" \
      -H "Content-Type: application/json" \
      -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
      -d "$PAYLOAD" \
      -w "\nHTTP_STATUS:%{http_code}")

    # Parse response
    HTTP_STATUS=$(echo "$RESPONSE" | tail -n1 | cut -d: -f2)
    BODY=$(echo "$RESPONSE" | head -n -1)

    echo "HTTP Status: $HTTP_STATUS"
    echo "Response: $BODY"
    
    if [ "$HTTP_STATUS" = "200" ]; then
        echo "✅ Test PASSED for $repo_name repository"
    else
        echo "❌ Test FAILED for $repo_name repository"
    fi
    
    echo ""
    sleep 2
}

# Test posts repository
test_webhook "posts" "jgndev/posts" "new-test-post.md"

# Test cheatsheets repository  
test_webhook "cheatsheets" "jgndev/cheatsheets" "new-test-cheatsheet.md"

# Test unknown repository (should trigger fallback)
test_webhook "unknown-repo" "jgndev/unknown-repo" "test-file.md"

echo "========================================="
echo "All webhook tests completed!"
echo "========================================="
echo "Check your server logs to verify:"
echo "1. Posts repository triggered ContentManager refresh"
echo "2. Cheatsheets repository triggered CheatsheetManager refresh"  
echo "3. Unknown repository triggered fallback (both managers)"
echo ""
echo "Expected log messages:"
echo "- 'Detected posts repository, refreshing ContentManager'"
echo "- 'Detected cheatsheets repository, refreshing CheatsheetManager'"
echo "- 'WARNING: Unknown repository 'unknown-repo', refreshing both managers as fallback'"