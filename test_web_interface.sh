#!/bin/bash

# Colors for better output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Function to print headers
print_header() {
  echo -e "\n${BLUE}==========================================${NC}"
  echo -e "${BLUE}$1${NC}"
  echo -e "${BLUE}==========================================${NC}\n"
}

# Generate random user ID for testing
USER_ID="test_user_$(date +%s)"
echo "Using test user ID: $USER_ID"

# 1. Test save settings endpoint
print_header "Test 1: Saving user settings"
echo "Sending request to save user settings..."
curl -X POST -H "Content-Type: application/json" \
  -d "{\"userId\":\"$USER_ID\",\"targetBloodSugar\":5.5,\"insulinSensitivityFactor\":2.0,\"insulinToCarbRatio\":10.0,\"baseInsulinCoefficient\":{\"morning\":1.0,\"afternoon\":0.8,\"evening\":1.2,\"night\":0.9}}" \
  $BASE_URL/api/settings

# 2. Test blood sugar endpoint
print_header "Test 2: Saving blood sugar reading"
echo "Sending request to save a blood sugar reading..."
curl -X POST -H "Content-Type: application/json" \
  -d "{\"userId\":\"$USER_ID\",\"value\":6.5}" \
  $BASE_URL/api/bloodsugar

# 3. Test food analysis endpoint
print_header "Test 3: Analyzing food with text only"
echo "Sending request for food analysis..."
curl -X POST -H "Content-Type: application/json" \
  -d "{\"userId\":\"$USER_ID\",\"food\":\"pasta with cheese and tomato sauce\"}" \
  $BASE_URL/api/analyze-food

# 4. Test static file serving
print_header "Test 4: Checking static files"
echo "Checking index.html..."
curl -I $BASE_URL/index.html

echo "Checking app.js..."
curl -I $BASE_URL/app.js

echo "Checking styles.css..."
curl -I $BASE_URL/styles.css

# Final message
print_header "Web interface testing completed"
echo "Open http://localhost:8080 in your browser to test the web interface manually."
echo "Use the test user ID: $USER_ID" 