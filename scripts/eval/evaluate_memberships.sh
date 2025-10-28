#!/bin/bash

# Simple bash script to evaluate memberships for a list of keycloak IDs
# Usage: ./evaluate_memberships.sh <bearer_token> [keycloak_ids_file] [api_base_url]

set -e

# Check arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 <bearer_token> [keycloak_ids_file] [api_base_url]"
    echo ""
    echo "Example:"
    echo "  $0 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...'"
    echo "  $0 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...' my_keycloak_ids.txt"
    echo "  $0 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...' my_keycloak_ids.txt https://api.yourdomain.com"
    exit 1
fi

# Default values
BEARER_TOKEN=$1
KEYCLOAK_IDS_FILE=${2:-"keycloak_ids.txt"}
API_BASE_URL=${3:-"https://api.kli.one/profile"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Membership Evaluation Script${NC}"
echo "================================"
echo "Keycloak IDs file: $KEYCLOAK_IDS_FILE"
echo "API Base URL: $API_BASE_URL"
echo "Bearer Token: ${BEARER_TOKEN:0:20}..." # Show first 20 chars of token
echo ""

# Check if the keycloak IDs file exists
if [ ! -f "$KEYCLOAK_IDS_FILE" ]; then
    echo -e "${RED}Error: Keycloak IDs file '$KEYCLOAK_IDS_FILE' not found${NC}"
    echo "Usage: $0 <bearer_token> [keycloak_ids_file] [api_base_url]"
    echo ""
    echo "Create a text file with one keycloak ID per line:"
    echo "12345678-1234-1234-1234-123456789012"
    echo "87654321-4321-4321-4321-210987654321"
    echo ""
    echo "Or specify a different file:"
    echo "  $0 '$BEARER_TOKEN' my_custom_file.txt"
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

# Count total lines (excluding comments and empty lines)
TOTAL_LINES=$(grep -v '^#' "$KEYCLOAK_IDS_FILE" | grep -v '^[[:space:]]*$' | wc -l | tr -d ' ')

if [ "$TOTAL_LINES" -eq 0 ]; then
    echo -e "${YELLOW}No keycloak IDs found in file (excluding comments and empty lines)${NC}"
    exit 0
fi

echo "Found $TOTAL_LINES keycloak IDs to evaluate"
echo ""

# Initialize counters
SUCCESS_COUNT=0
ERROR_COUNT=0
CURRENT=0

# Read file line by line
while IFS= read -r line || [ -n "$line" ]; do
    # Skip comments and empty lines
    if [[ "$line" =~ ^[[:space:]]*# ]] || [[ -z "${line// }" ]]; then
        continue
    fi
    
    # Trim whitespace
    keycloak_id=$(echo "$line" | xargs)
    
    if [ -z "$keycloak_id" ]; then
        continue
    fi
    
    CURRENT=$((CURRENT + 1))
    
    echo -n "[$CURRENT/$TOTAL_LINES] Evaluating $keycloak_id... "
    
    # Create JSON payload
    json_payload=$(cat <<EOF
{
  "keycloak_id": "$keycloak_id"
}
EOF
)
    
    # Make HTTP request
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $BEARER_TOKEN" \
        -d "$json_payload" \
        "$API_BASE_URL/v1/membership/evaluation" 2>/dev/null)
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    
    # Extract response body (all lines except last) - portable approach
    response_body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq 200 ]; then
        echo -e "${GREEN}SUCCESS${NC}"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        
        # Optionally show response details
        if [ "$VERBOSE" = "1" ]; then
            echo "  Response: $response_body"
        fi
    else
        echo -e "${RED}ERROR (HTTP $http_code)${NC}"
        ERROR_COUNT=$((ERROR_COUNT + 1))
        
        # Show error details
        echo "  Error: $response_body"
    fi
    
    # Small delay to avoid overwhelming the server
    sleep 0.1
    
done < "$KEYCLOAK_IDS_FILE"

echo ""
echo "================================"
echo -e "${GREEN}Success: $SUCCESS_COUNT${NC}"
echo -e "${RED}Errors: $ERROR_COUNT${NC}"
echo "Total processed: $CURRENT"

if [ $ERROR_COUNT -gt 0 ]; then
    exit 1
else
    echo -e "${GREEN}All evaluations completed successfully!${NC}"
fi
