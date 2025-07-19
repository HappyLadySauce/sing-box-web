# API Testing Scripts

This directory contains scripts for testing the sing-box-web backend API.

## Files

- `test_api.py` - Comprehensive API testing script
- `main.py` - Original simple test script (legacy)

## Usage

### Prerequisites

Install required Python packages:
```bash
pip install requests
```

### Running the Tests

#### Basic Usage
```bash
# Test against local server (default: http://localhost:8080)
python scripts/test_api.py

# Test against custom server
python scripts/test_api.py --url http://your-server:8080

# Enable verbose output
python scripts/test_api.py --verbose
```

#### Example Commands
```bash
# Test local development server
python scripts/test_api.py

# Test production server with verbose output
python scripts/test_api.py --url https://api.yoursite.com --verbose

# Test staging server
python scripts/test_api.py --url http://staging.yoursite.com:8080
```

## Test Coverage

The test script covers the following API endpoints:

### Health Checks
- `/health` - Service health status
- `/livez` - Liveness probe
- `/readyz` - Readiness probe

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/profile` - Get user profile
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout

### User Management (Admin)
- `GET /api/v1/admin/users` - List users
- `POST /api/v1/admin/users` - Create user
- `GET /api/v1/admin/users/{id}` - Get user details
- `PUT /api/v1/admin/users/{id}` - Update user
- `DELETE /api/v1/admin/users/{id}` - Delete user
- `POST /api/v1/admin/users/{id}/reset-traffic` - Reset user traffic
- `GET /api/v1/admin/users/{id}/nodes` - Get user nodes
- `POST /api/v1/admin/users/{id}/nodes/{node_id}` - Add user to node
- `DELETE /api/v1/admin/users/{id}/nodes/{node_id}` - Remove user from node

### Node Management (Admin)
- `GET /api/v1/admin/nodes` - List nodes
- `POST /api/v1/admin/nodes` - Create node
- `GET /api/v1/admin/nodes/{id}` - Get node details
- `PUT /api/v1/admin/nodes/{id}` - Update node
- `DELETE /api/v1/admin/nodes/{id}` - Delete node
- `GET /api/v1/admin/nodes/{id}/users` - Get node users

### Plan Management (Admin)
- `GET /api/v1/admin/plans` - List plans
- `POST /api/v1/admin/plans` - Create plan
- `GET /api/v1/admin/plans/{id}` - Get plan details
- `PUT /api/v1/admin/plans/{id}` - Update plan
- `DELETE /api/v1/admin/plans/{id}` - Delete plan

### Traffic Statistics
- `GET /api/v1/traffic/statistics` - Get traffic statistics
- `GET /api/v1/traffic/chart` - Get traffic chart data
- `GET /api/v1/traffic/live` - Get live traffic data
- `GET /api/v1/traffic/summary` - Get traffic summary
- `GET /api/v1/traffic/top-users` - Get top users by traffic
- `GET /api/v1/traffic/top-nodes` - Get top nodes by traffic

### System Monitoring
- `GET /api/v1/system/status` - System status
- `GET /api/v1/system/dashboard` - Dashboard data
- `GET /api/v1/system/statistics` - System statistics

## Test Results

The script generates:
1. Console output with test results
2. JSON file with detailed results (format: `api_test_results_YYYYMMDD_HHMMSS.json`)

### Sample Output
```
========================================
API TEST SUMMARY
========================================
Total Tests: 45
Passed: 42
Failed: 3
Success Rate: 93.3%
Duration: 12.45 seconds
========================================

Detailed results saved to: api_test_results_20240119_143022.json
```

## Authentication

The test script uses the following authentication flow:
1. Attempts to register a new user
2. Falls back to admin login (username: `admin`, password: `admin123`)
3. Uses JWT tokens for authenticated requests
4. Tests token refresh functionality

## Test Data

The script creates temporary test data:
- Test users with unique timestamps
- Test nodes with example configurations
- Test plans with sample pricing
- All test data is automatically cleaned up after tests complete

## Error Handling

The script includes comprehensive error handling:
- Network timeouts and connection errors
- Invalid response formats
- Authentication failures
- Server errors (5xx status codes)
- Invalid request data (4xx status codes)

## Extending the Tests

To add new test cases:

1. Add a new test method to the `SingBoxAPITester` class:
```python
def test_new_feature(self):
    """Test new feature endpoints"""
    logger.info("Testing new feature...")
    
    response = self.make_request("GET", "/api/v1/new-endpoint")
    if response and response.status_code == 200:
        self.log_test("New Feature Test", True, "Successfully tested new feature")
    else:
        status_code = response.status_code if response else "No response"
        self.log_test("New Feature Test", False, f"Status: {status_code}")
```

2. Call the new test method in `run_all_tests()`:
```python
def run_all_tests(self):
    # ... existing tests ...
    self.test_new_feature()
    # ... cleanup ...
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure the API server is running
   - Check the URL and port
   - Verify firewall settings

2. **Authentication Failures**
   - Check if admin credentials are correct
   - Verify JWT implementation is working
   - Check token expiration settings

3. **Test Failures**
   - Check server logs for error details
   - Verify database connections
   - Ensure required dependencies are installed

### Debug Mode

Run with verbose output to see detailed request/response information:
```bash
python scripts/test_api.py --verbose
```

This will show:
- All test results (pass/fail)
- HTTP status codes
- Response details
- Timing information