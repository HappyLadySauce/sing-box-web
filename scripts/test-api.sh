#!/bin/bash

# API Test Script for sing-box-web
# This script tests the main API endpoints

BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
    fi
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Function to test an endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    local auth_header=$5
    
    print_info "Testing: $description"
    
    if [ -n "$data" ]; then
        if [ -n "$auth_header" ]; then
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $auth_header" \
                -d "$data" \
                "$API_BASE$endpoint")
        else
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$API_BASE$endpoint")
        fi
    else
        if [ -n "$auth_header" ]; then
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" \
                -H "Authorization: Bearer $auth_header" \
                "$API_BASE$endpoint")
        else
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" \
                "$API_BASE$endpoint")
        fi
    fi
    
    # Extract status code
    status_code=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo "$response" | sed -e 's/HTTPSTATUS\:.*//g')
    
    # Check if successful (2xx status codes)
    if [[ $status_code -ge 200 && $status_code -lt 300 ]]; then
        print_status 0 "$description - Status: $status_code"
        return 0
    else
        print_status 1 "$description - Status: $status_code"
        echo "Response: $body"
        return 1
    fi
}

# Start testing
echo "======================================"
echo "🚀 sing-box-web API Test Suite"
echo "======================================"
echo ""

# Test 1: Health check endpoints
print_info "Testing health check endpoints..."
test_endpoint "GET" "/health" "Health check"
test_endpoint "GET" "/livez" "Liveness check"
test_endpoint "GET" "/readyz" "Readiness check"
echo ""

# Test 2: Authentication endpoints
print_info "Testing authentication endpoints..."

# Test login with invalid credentials (should fail)
test_endpoint "POST" "/auth/login" "Login with invalid credentials" \
    '{"username": "invalid", "password": "invalid"}'

# Test login with demo credentials
login_response=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin123"}' \
    "$API_BASE/auth/login")

# Extract access token if login successful
access_token=$(echo "$login_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$access_token" ]; then
    print_status 0 "Login with demo credentials - Token received"
    echo ""
    
    # Test authenticated endpoints
    print_info "Testing authenticated endpoints..."
    
    # Test profile endpoint
    test_endpoint "GET" "/auth/profile" "Get user profile" "" "$access_token"
    
    # Test system endpoints
    test_endpoint "GET" "/system/status" "System status" "" "$access_token"
    test_endpoint "GET" "/system/dashboard" "System dashboard" "" "$access_token"
    test_endpoint "GET" "/system/statistics" "System statistics" "" "$access_token"
    test_endpoint "GET" "/system/health" "System health" "" "$access_token"
    
    # Test traffic endpoints
    test_endpoint "GET" "/traffic/statistics" "Traffic statistics" "" "$access_token"
    test_endpoint "GET" "/traffic/summary" "Traffic summary" "" "$access_token"
    test_endpoint "GET" "/traffic/live" "Live traffic" "" "$access_token"
    
    echo ""
    
    # Test admin endpoints
    print_info "Testing admin endpoints..."
    
    # Test user management endpoints
    test_endpoint "GET" "/admin/users" "List users" "" "$access_token"
    test_endpoint "GET" "/admin/users/1" "Get user by ID" "" "$access_token"
    
    # Test node management endpoints
    test_endpoint "GET" "/admin/nodes" "List nodes" "" "$access_token"
    
    # Test create user (should work)
    test_endpoint "POST" "/admin/users" "Create user" \
        '{"username": "testuser", "email": "test@example.com", "password": "password123", "plan_id": 1}' \
        "$access_token"
    
    # Test logout
    test_endpoint "POST" "/auth/logout" "Logout" "" "$access_token"
    
else
    print_status 1 "Login with demo credentials - No token received"
    echo "Response: $login_response"
fi

echo ""

# Test 3: Public endpoints that don't require authentication
print_info "Testing public endpoints..."
test_endpoint "GET" "/auth/login" "Login endpoint availability (GET should not be allowed)"

echo ""
echo "======================================"
echo "🏁 Test Summary"
echo "======================================"
echo ""
echo "Test completed! Check the output above for any failures."
echo ""
echo "Next steps:"
echo "1. Start the application: ./bin/sing-box-web web --config configs/web.yaml"
echo "2. Run this test script: ./scripts/test-api.sh"
echo "3. Check the logs for any errors"
echo ""
echo "Available API endpoints:"
echo "• Public:"
echo "  - POST /api/v1/auth/login"
echo "  - POST /api/v1/auth/refresh"
echo "• Authenticated:"
echo "  - GET /api/v1/auth/profile"
echo "  - POST /api/v1/auth/logout"
echo "  - GET /api/v1/system/*"
echo "  - GET /api/v1/traffic/*"
echo "• Admin:"
echo "  - /api/v1/admin/users/*"
echo "  - /api/v1/admin/nodes/*"
echo ""