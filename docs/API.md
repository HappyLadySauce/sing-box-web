# sing-box-web API Documentation

## Overview

The sing-box-web API provides a RESTful interface for managing a distributed sing-box management platform. The API supports user authentication, node management, traffic statistics, and system monitoring.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

The API uses JWT (JSON Web Token) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <token>
```

## Endpoints

### Public Endpoints

#### Authentication

##### Login
```http
POST /auth/login
```

Request body:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "1",
    "username": "admin",
    "role": "admin",
    "email": "admin@localhost"
  }
}
```

##### Refresh Token
```http
POST /auth/refresh
```

Request body:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Authenticated Endpoints

#### User Profile

##### Get Profile
```http
GET /auth/profile
```

##### Logout
```http
POST /auth/logout
```

#### System Monitoring

##### System Status
```http
GET /system/status
```

Response:
```json
{
  "data": {
    "status": "running",
    "version": "1.0.0",
    "uptime": "2h30m",
    "database_status": "healthy",
    "statistics": {
      "total_users": 150,
      "active_users": 120,
      "total_nodes": 25,
      "online_nodes": 20,
      "total_traffic": 1048576000,
      "today_traffic": 52428800,
      "monthly_traffic": 838860800
    }
  }
}
```

##### Dashboard Data
```http
GET /system/dashboard
```

##### System Statistics
```http
GET /system/statistics
```

##### Health Check
```http
GET /system/health
```

#### Traffic Statistics

##### Traffic Statistics
```http
GET /traffic/statistics
```

Query parameters:
- `start_date` (optional): Start date in YYYY-MM-DD format
- `end_date` (optional): End date in YYYY-MM-DD format
- `user_id` (optional): Filter by user ID
- `node_id` (optional): Filter by node ID
- `granularity` (optional): hourly, daily, monthly (default: daily)
- `include_top` (optional): Include top users/nodes (default: false)

##### Traffic Chart
```http
GET /traffic/chart
```

##### Live Traffic
```http
GET /traffic/live
```

##### Traffic Summary
```http
GET /traffic/summary
```

##### Top Users
```http
GET /traffic/top-users
```

##### Top Nodes
```http
GET /traffic/top-nodes
```

### Admin Endpoints

#### User Management

##### List Users
```http
GET /admin/users
```

Query parameters:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)
- `search` (optional): Search query
- `status` (optional): Filter by status (active, suspended, expired, disabled)
- `plan_id` (optional): Filter by plan ID

##### Create User
```http
POST /admin/users
```

Request body:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123",
  "display_name": "New User",
  "plan_id": 1,
  "traffic_quota": 10737418240,
  "device_limit": 3,
  "speed_limit": 0
}
```

##### Get User
```http
GET /admin/users/{id}
```

##### Update User
```http
PUT /admin/users/{id}
```

##### Delete User
```http
DELETE /admin/users/{id}
```

##### Reset User Traffic
```http
POST /admin/users/{id}/reset-traffic
```

##### Get User Nodes
```http
GET /admin/users/{id}/nodes
```

##### Add User to Node
```http
POST /admin/users/{id}/nodes/{node_id}
```

##### Remove User from Node
```http
DELETE /admin/users/{id}/nodes/{node_id}
```

#### Node Management

##### List Nodes
```http
GET /admin/nodes
```

Query parameters:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)
- `search` (optional): Search query
- `type` (optional): Filter by node type
- `status` (optional): Filter by status
- `region` (optional): Filter by region
- `enabled` (optional): Filter by enabled status

##### Create Node
```http
POST /admin/nodes
```

Request body:
```json
{
  "name": "Node 1",
  "description": "Primary node in US",
  "type": "vmess",
  "host": "node1.example.com",
  "port": 443,
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "region": "US",
  "country": "United States",
  "city": "New York",
  "max_users": 100,
  "speed_limit": 0,
  "traffic_rate": 1.0,
  "is_enabled": true
}
```

##### Get Node
```http
GET /admin/nodes/{id}
```

##### Update Node
```http
PUT /admin/nodes/{id}
```

##### Delete Node
```http
DELETE /admin/nodes/{id}
```

##### Enable Node
```http
POST /admin/nodes/{id}/enable
```

##### Disable Node
```http
POST /admin/nodes/{id}/disable
```

##### Update Node Heartbeat
```http
POST /admin/nodes/{id}/heartbeat
```

##### Update Node System Info
```http
POST /admin/nodes/{id}/system-info
```

Request body:
```json
{
  "cpu_usage": 45.5,
  "memory_usage": 60.2,
  "disk_usage": 75.8,
  "load1": 2.1,
  "load5": 1.8,
  "load15": 1.5
}
```

##### Get Node Users
```http
GET /admin/nodes/{id}/users
```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request format",
  "details": "validation error details"
}
```

### 401 Unauthorized
```json
{
  "error": "Invalid username or password"
}
```

### 403 Forbidden
```json
{
  "error": "Access denied"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

## Data Models

### User
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@localhost",
  "display_name": "Administrator",
  "status": "active",
  "plan_id": 1,
  "plan_name": "Premium Plan",
  "traffic_quota": 10737418240,
  "traffic_used": 1073741824,
  "traffic_remaining": 9663676416,
  "device_limit": 5,
  "speed_limit": 0,
  "expires_at": "2024-12-31T23:59:59Z",
  "last_login_at": "2024-01-15T10:30:00Z",
  "last_login_ip": "192.168.1.100",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Node
```json
{
  "id": 1,
  "name": "Node 1",
  "description": "Primary node in US",
  "type": "vmess",
  "status": "online",
  "host": "node1.example.com",
  "port": 443,
  "region": "US",
  "country": "United States",
  "city": "New York",
  "max_users": 100,
  "current_users": 45,
  "usage_percentage": 45.0,
  "speed_limit": 0,
  "traffic_rate": 1.0,
  "total_traffic": 5368709120,
  "upload_traffic": 2147483648,
  "download_traffic": 3221225472,
  "cpu_usage": 45.5,
  "memory_usage": 60.2,
  "disk_usage": 75.8,
  "is_enabled": true,
  "is_online": true,
  "last_heartbeat": "2024-01-15T10:29:00Z",
  "agent_version": "1.0.0",
  "sing_box_version": "1.5.0",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-15T10:29:00Z"
}
```

### Traffic Statistics
```json
{
  "total_upload": 2147483648,
  "total_download": 3221225472,
  "total_traffic": 5368709120,
  "start_date": "2024-01-01",
  "end_date": "2024-01-15",
  "chart_data": [
    {
      "date": "2024-01-01",
      "upload": 104857600,
      "download": 157286400,
      "total": 262144000
    }
  ]
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse:
- 100 requests per minute per IP address
- 1000 requests per hour per authenticated user

## Testing

Use the provided test script to verify API functionality:

```bash
# Make the script executable
chmod +x scripts/test-api.sh

# Run the test
./scripts/test-api.sh
```

## Support

For API support and questions, please refer to the project documentation or create an issue in the project repository.