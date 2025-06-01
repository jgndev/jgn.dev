# GitHub Token Setup Guide

This guide explains how to create and configure a GitHub Personal Access Token for stratocraft.dev to access your posts repository and avoid API rate limiting.

## Why Do You Need a GitHub Token?

Without authentication, GitHub limits API requests to **60 per hour**. With a Personal Access Token, you get **5,000 requests per hour**, which is essential for:

- Fetching posts from your GitHub repository
- Avoiding rate limiting during development
- Enabling webhook functionality for automatic updates
- Accessing private repositories (if needed)

## Creating a Personal Access Token

### Step 1: Navigate to GitHub Settings

1. **Log in to GitHub** and click your profile picture in the top-right corner
2. **Select "Settings"** from the dropdown menu
3. **Scroll down** to "Developer settings" in the left sidebar
4. **Click "Personal access tokens"**
5. **Choose "Tokens (classic)"** for compatibility

### Step 2: Generate New Token

1. **Click "Generate new token"** → **"Generate new token (classic)"**
2. **Enter a descriptive note** like `stratocraft.dev blog application`
3. **Set expiration** (recommended: 90 days or longer for production)
4. **Select the required scopes** (see permissions section below)

### Step 3: Required Permissions

Select **only the minimal permissions needed**:

#### For Public Repositories:
- ✅ **`public_repo`** - Access public repositories
  - This allows reading your public posts repository

#### For Private Repositories:
- ✅ **`repo`** - Full control of private repositories
  - Only needed if your posts repository is private

#### No Additional Permissions Needed
- ❌ **`workflow`** - Not needed
- ❌ **`write:packages`** - Not needed  
- ❌ **`delete:packages`** - Not needed
- ❌ **`admin:repo_hook`** - Not needed (webhooks are configured separately)

### Step 4: Generate and Copy Token

1. **Click "Generate token"** at the bottom
2. **Copy the token immediately** - GitHub will only show it once
3. **Store it securely** - treat it like a password

## Configuring the Token

### For Development

Set the token as an environment variable:

**macOS/Linux:**
```bash
export GITHUB_TOKEN=ghp_your_token_here

# Add to your shell profile for persistence
echo 'export GITHUB_TOKEN=ghp_your_token_here' >> ~/.zshrc
source ~/.zshrc
```

**Windows (PowerShell):**
```powershell
$env:GITHUB_TOKEN="ghp_your_token_here"

# For persistence, add to your PowerShell profile
Add-Content $PROFILE '$env:GITHUB_TOKEN="ghp_your_token_here"'
```

### For Production (Azure Deployment)

The deployment scripts will automatically use the `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=ghp_your_token_here
./scripts/deploy-azure-appservice.sh
```

## Token Security Best Practices

### ✅ Do's

- **Use descriptive names** when creating tokens
- **Set reasonable expiration dates** (30-90 days for development, longer for production)
- **Use minimal required permissions** only
- **Store tokens as environment variables**, never in code
- **Regenerate tokens periodically**
- **Delete unused tokens** immediately

### ❌ Don'ts

- **Never commit tokens to Git repositories**
- **Don't share tokens** in chat, email, or documentation
- **Avoid setting "No expiration"** unless absolutely necessary
- **Don't grant excessive permissions** (like admin access)

## Troubleshooting

### Invalid Token Error

```
Error: 401 Unauthorized - Bad credentials
```

**Solutions:**
1. **Verify token is correctly set**: `echo $GITHUB_TOKEN`
2. **Check token hasn't expired** in GitHub Settings
3. **Ensure token has correct permissions** (public_repo or repo)
4. **Try regenerating the token** if it's corrupted

### Rate Limiting Still Occurring

```
Error: 403 Forbidden - API rate limit exceeded
```

**Solutions:**
1. **Confirm token is being used**: Check application logs
2. **Verify environment variable is set**: `echo $GITHUB_TOKEN`
3. **Check token permissions** include repository access
4. **Wait for rate limit reset** (resets every hour)

### Permission Denied

```
Error: 403 Forbidden - Repository access blocked
```

**Solutions:**
1. **For private repos**: Ensure token has `repo` scope
2. **For public repos**: Ensure token has `public_repo` scope
3. **Verify repository exists** and is accessible
4. **Check organization permissions** if applicable

### Token Not Found in Environment

```
Warning: GITHUB_TOKEN environment variable is not set
```

**Solutions:**
1. **Set the environment variable**:
   ```bash
   export GITHUB_TOKEN=ghp_your_token_here
   ```
2. **Restart your terminal** to load new environment variables
3. **Check your shell profile** (.zshrc, .bashrc, etc.) for persistence

## Testing Your Token

### Quick Test

Verify your token works with a simple curl command:

```bash
curl -H "Authorization: token $GITHUB_TOKEN" \
  https://api.github.com/user
```

**Expected response**: Your GitHub user information in JSON format.

### Application Test

Run the stratocraft.dev application and check logs for:

```
✅ Successfully authenticated with GitHub API
✅ Rate limit: 5000 requests per hour
```

## Token Management

### Viewing Your Tokens

1. **Go to GitHub Settings** → **Developer settings** → **Personal access tokens**
2. **Review active tokens** and their last usage
3. **Delete unused tokens** regularly

### Regenerating Tokens

1. **Click on the token name** in your token list
2. **Click "Regenerate token"**
3. **Update the environment variable** with the new token
4. **Restart your application**

### Token Expiration

When your token expires:

1. **Create a new token** following the same steps
2. **Update your environment variables**
3. **Update production deployments** with the new token
4. **Delete the expired token** from GitHub

## Repository Configuration

### Posts Repository Setup

Your GitHub token needs access to the repository containing your blog posts:

1. **Repository must be accessible** with your token permissions
2. **Posts should be in Markdown format** with YAML frontmatter
3. **Repository structure** should match your site configuration

### Example Repository Structure

```
your-posts-repo/
├── post-1.md
├── post-2.md
├── drafts/
│   └── upcoming-post.md
└── README.md
```

## Integration with stratocraft.dev

### Environment Variable Usage

The application automatically detects and uses your GitHub token:

```go
token := os.Getenv("GITHUB_TOKEN")
if token != "" {
    // Use authenticated requests (5,000/hour limit)
} else {
    // Use unauthenticated requests (60/hour limit)
}
```

### Rate Limit Monitoring

Monitor your usage in the application logs:

```
INFO: GitHub API rate limit: 4,832/5,000 remaining
INFO: Rate limit resets at: 2024-01-15T10:30:00Z
```

## Support and Resources

### Official Documentation

- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [GitHub API Rate Limiting](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)
- [GitHub API Authentication](https://docs.github.com/en/rest/guides/getting-started-with-the-rest-api#authentication)

### Common Use Cases

**Blog Development:**
- Repository access: `public_repo` or `repo`
- Typical usage: 50-200 API calls per session
- Recommended expiration: 90 days

**Production Deployment:**
- Repository access: `public_repo` or `repo`
- Typical usage: 10-50 API calls per day
- Recommended expiration: 1 year (with monitoring)

### Getting Help

If you encounter issues:

1. **Check the troubleshooting section** above
2. **Review GitHub's status page** for API issues
3. **Open an issue** in the stratocraft.dev repository with error logs
4. **Include relevant error messages** but never include your actual token 