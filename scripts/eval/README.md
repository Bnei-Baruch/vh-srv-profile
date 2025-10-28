# Membership Evaluation Script

A simple bash script to evaluate memberships for a list of keycloak IDs by calling the production API endpoint.

## Usage

1. Get your Bearer token from your authentication system (Keycloak, etc.)

2. Create a text file with keycloak IDs (one per line) - optional, defaults to `keycloak_ids.txt`:
   ```bash
   # Create your keycloak IDs file
   cat > keycloak_ids.txt << EOF
   12345678-1234-1234-1234-123456789012
   87654321-4321-4321-4321-210987654321
   11111111-2222-3333-4444-555555555555
   EOF
   ```

3. Run the script:
   ```bash
   # Basic usage (uses default file and API URL)
   ./evaluate_memberships.sh "your_bearer_token_here"
   
   # With custom keycloak IDs file
   ./evaluate_memberships.sh "your_bearer_token_here" my_keycloak_ids.txt
   
   # With custom file and API URL
   ./evaluate_memberships.sh "your_bearer_token_here" my_keycloak_ids.txt https://api.yourdomain.com
   ```

4. For verbose output (shows response details):
   ```bash
   VERBOSE=1 ./evaluate_memberships.sh "your_bearer_token_here"
   ```

## Features

- ✅ Simple to use - just provide a Bearer token (file and URL are optional)
- ✅ Bearer token authentication support
- ✅ Sensible defaults (keycloak_ids.txt file, configurable API URL)
- ✅ Colored output for easy reading
- ✅ Progress tracking
- ✅ Error handling and reporting
- ✅ Skips comments and empty lines
- ✅ Rate limiting (small delay between requests)
- ✅ Summary of results

## File Format

The keycloak IDs file should contain one keycloak ID per line:
```
12345678-1234-1234-1234-123456789012
87654321-4321-4321-4321-210987654321
# This is a comment and will be ignored
11111111-2222-3333-4444-555555555555
```

## Requirements

- `curl` command-line tool
- `bash` shell
- Network access to the production API
- Valid Bearer token for API authentication

## Example Output

```
Membership Evaluation Script
================================
Keycloak IDs file: keycloak_ids.txt
API Base URL: https://api.yourdomain.com
Bearer Token: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

Found 3 keycloak IDs to evaluate

[1/3] Evaluating 12345678-1234-1234-1234-123456789012... SUCCESS
[2/3] Evaluating 87654321-4321-4321-4321-210987654321... SUCCESS
[3/3] Evaluating 11111111-2222-3333-4444-555555555555... ERROR (HTTP 404)

================================
Success: 2
Errors: 1
Total processed: 3
```

## Quick Start

The simplest way to get started:

1. **Get your token** from Keycloak or your auth system
2. **Create the default file**:
   ```bash
   echo "your-keycloak-id-here" > keycloak_ids.txt
   ```
3. **Run the script**:
   ```bash
   ./evaluate_memberships.sh "your_bearer_token_here"
   ```

That's it! The script will use `keycloak_ids.txt` as the default file and the configured API URL.

## Getting Your Bearer Token

You'll need to obtain a Bearer token from your authentication system. This is typically done through:

1. **Keycloak Admin Console** - Get a token from your realm
2. **API Authentication** - Use your client credentials
3. **User Login** - Authenticate as a user and get their token

The token should be a JWT (JSON Web Token) that starts with `eyJ` and contains the necessary permissions to access the membership evaluation endpoint.
