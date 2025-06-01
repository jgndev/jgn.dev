# GitHub Webhook Setup Guide

This guide explains how to set up GitHub webhooks for stratocraft.dev site automatically refreshes content when new posts are added to your posts repository.

## Overview

When you push new `.md` files (blog posts) to your posts repository, GitHub will automatically notify your website, which will then refresh its content without requiring a restart.

## Prerequisites

1. Your website must be publicly accessible (not just `localhost`)
2. You need admin access to your posts repository on GitHub
3. You need access to set environment variables on your server

## Step 1: Generate a Webhook Secret

First, generate a secure secret that GitHub will use to sign webhook requests:

```bash
openssl rand -hex 32
```

Save this secret - you'll need it in both your server environment and GitHub webhook configuration.

## Step 2: Configure Your Server

Set the webhook secret as an environment variable on your server:

```bash
export GITHUB_WEBHOOK_SECRET=your_generated_secret_here
```

For production deployments, add this to your deployment configuration or `.env` file. Reminder that `.env` files should
not be checked into source control and should be added to your `.gitignore`.

## Step 3: Configure GitHub Webhook

1. Go to your posts repository on GitHub
2. Navigate to **Settings** → **Webhooks**
3. Click **Add webhook**
4. Fill in the webhook details:
   - **Payload URL**: 
     - For App Service: `https://your-app-name.azurewebsites.net/webhook/github`
     - For custom domain: `https://yourdomain.com/webhook/github`
   - **Content type**: `application/json`
   - **Secret**: The secret you generated in Step 1
   - **Events**: Select "Just the push event"
   - **Active**: ✅ Checked

5. Click **Add webhook**  

💡 **App Service Note**: Your webhook URL will be `https://stratocraft-webapp.azurewebsites.net/webhook/github` if you used the default app name, or replace `stratocraft-webapp` with your chosen app name.

💡 If your site is not deployed yet, you will get a message `Last delivery was not successful. Failed to connect to host.` This is expected behavior until your deployment is completed and the URL resolves.

## Step 4: Test the Webhook

1. Make sure your website is running with the webhook secret configured
2. Add or modify a `.md` file in your posts repository
3. Commit and push the changes to the main branch
4. Check your server logs - you should see messages like:
   ```
   Detected markdown file change: new-post.md
   Refreshing content due to webhook from username/posts-repo
   Successfully refreshed content from webhook
   ```

## How It Works

1. **GitHub Push**: When you push changes to your posts repo
2. **Webhook Trigger**: GitHub sends a POST request to `/webhook/github`
3. **Signature Verification**: Your server verifies the request came from GitHub
4. **Change Detection**: Server checks if any `.md` files were added/modified
5. **Content Refresh**: If markdown files changed, server calls `ContentManager.RefreshContent()`
6. **Live Update**: New posts are immediately available to readers

## Security Features

- **Signature Verification**: Uses HMAC-SHA256 to verify requests came from GitHub
- **Branch Filtering**: Only responds to pushes to main/master branch
- **File Type Filtering**: Only triggers refresh when markdown files are changed
- **Environment Isolation**: Webhook secret is stored as environment variable

## Troubleshooting

### Webhook Not Triggering

1. Check your server logs for error messages
2. Verify the webhook secret matches between GitHub and your server
3. Ensure your webhook URL is publicly accessible
4. Check GitHub's webhook delivery logs in the repo settings

### Content Not Refreshing

1. Verify the push was to the main/master branch
2. Confirm `.md` files were actually added/modified in the commit
3. Check server logs for `RefreshContent()` errors
4. Ensure your `GITHUB_TOKEN` is still valid

### Rate Limiting

- The webhook system respects your existing GitHub API rate limits
- With a `GITHUB_TOKEN`, you have 5,000 requests/hour
- Without authentication, you're limited to 60 requests/hour

## Production Considerations

- Use HTTPS for your webhook endpoint
- Consider implementing webhook delivery retry logic
- Monitor webhook delivery success rates
- Set up alerts for failed content refreshes
- Consider rate limiting the webhook endpoint to prevent abuse

## Testing Locally

For local development, you can use tools like [ngrok](https://ngrok.com/) to expose your local server:

```bash
# In one terminal
ngrok http 8080

# Use the ngrok URL as your webhook endpoint
# Example: https://abc123.ngrok.io/webhook/github
```

This allows you to test webhook functionality during development. 