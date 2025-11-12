# API Documentation

## Base URL

```
http://localhost:8080
```

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow this structure:

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": {}
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message"
}
```

## Endpoints

### Health & Monitoring

#### GET /health

Check if the service is running.

**Response**
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "timestamp": 1699891200,
    "version": "1.0.0"
  }
}
```

#### GET /ready

Check if the service and its dependencies are ready.

**Response**
```json
{
  "success": true,
  "message": "Service is ready",
  "data": {
    "status": "ready"
  }
}
```

---

### Authentication

#### POST /api/v1/auth/register

Register a new user.

**Request Body**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Validation Rules**
- `username`: Required, 3-50 characters
- `email`: Required, valid email format
- `password`: Required, minimum 6 characters

**Response (201 Created)**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-11-13T16:00:00Z",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "email": "john@example.com",
      "role": "user"
    }
  }
}
```

**Error Responses**
- `400 Bad Request`: Invalid input
- `409 Conflict`: Username or email already exists

---

#### POST /api/v1/auth/login

Authenticate a user and receive a JWT token.

**Request Body**
```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-11-13T16:00:00Z",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "email": "john@example.com",
      "role": "user"
    }
  }
}
```

**Error Responses**
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Invalid credentials
- `403 Forbidden`: Account is inactive

---

#### POST /api/v1/auth/refresh

Refresh an existing JWT token.

**Request Body**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK)**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-11-13T16:00:00Z"
  }
}
```

**Error Responses**
- `401 Unauthorized`: Invalid or expired token
- `404 Not Found`: User not found

---

### User Profile

#### GET /api/v1/profile

Get the authenticated user's profile.

**Headers**
```
Authorization: Bearer <token>
```

**Response (200 OK)**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "username": "john_doe",
    "email": "john@example.com",
    "role": "user"
  }
}
```

**Error Responses**
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: User not found

---

### Service Proxy

#### ANY /api/v1/*path

Proxy requests to backend services.

**Headers**
```
Authorization: Bearer <token>
```

**Example: GET /api/v1/users/123**

This request will be forwarded to the `users` service as `GET /123`

**Added Headers**
- `X-Forwarded-For`: Client IP address
- `X-Forwarded-Proto`: Request protocol
- `X-Forwarded-Host`: Original host

**Response**

The response from the backend service is returned as-is.

**Error Responses**
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Service not found
- `503 Service Unavailable`: Service temporarily unavailable (circuit breaker open)

---

### Admin - Service Management

#### GET /api/v1/admin/services

List all registered services.

**Headers**
```
Authorization: Bearer <admin-token>
```

**Response (200 OK)**
```json
{
  "success": true,
  "message": "Services retrieved successfully",
  "data": [
    {
      "name": "users",
      "urls": [
        "http://localhost:3001",
        "http://localhost:3002"
      ],
      "health_url": "/health",
      "active": true
    },
    {
      "name": "products",
      "urls": [
        "http://localhost:3003"
      ],
      "health_url": "/health",
      "active": true
    }
  ]
}
```

**Error Responses**
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Insufficient permissions (not admin)

---

#### POST /api/v1/admin/services

Register a new service.

**Headers**
```
Authorization: Bearer <admin-token>
```

**Request Body**
```json
{
  "name": "payments",
  "urls": [
    "http://localhost:3005",
    "http://localhost:3006"
  ],
  "health_url": "/health"
}
```

**Response (201 Created)**
```json
{
  "success": true,
  "message": "Service registered successfully",
  "data": null
}
```

**Error Responses**
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Insufficient permissions

---

#### DELETE /api/v1/admin/services/:name

Unregister a service.

**Headers**
```
Authorization: Bearer <admin-token>
```

**Parameters**
- `name`: Service name (path parameter)

**Response (200 OK)**
```json
{
  "success": true,
  "message": "Service unregistered successfully",
  "data": null
}
```

**Error Responses**
- `401 Unauthorized`: Missing or invalid token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Service not found

---

## Rate Limiting

All API endpoints are rate-limited. The following headers are included in responses:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1699891200
```

**Rate Limit Exceeded (429 Too Many Requests)**
```json
{
  "success": false,
  "error": "Rate limit exceeded. Please try again later."
}
```

The `Retry-After` header indicates when you can retry:
```
Retry-After: 30
```

---

## Error Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 503 | Service Unavailable - Service temporarily unavailable |

---

## Examples

### Complete Authentication Flow

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "secure123"
  }'

# Save the token from response
TOKEN="eyJhbGciOiJIUzI1NiIs..."

# 2. Access protected resource
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer $TOKEN"

# 3. Proxy to backend service
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN"
```

### Admin Operations

```bash
# Login as admin (you need to manually set role to 'admin' in database)
TOKEN="<admin-token>"

# List services
curl -X GET http://localhost:8080/api/v1/admin/services \
  -H "Authorization: Bearer $TOKEN"

# Register new service
curl -X POST http://localhost:8080/api/v1/admin/services \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "notifications",
    "urls": ["http://localhost:3007"],
    "health_url": "/health"
  }'
```

---

## Postman Collection

Import this collection to test the API:

```json
{
  "info": {
    "name": "API Gateway",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080"
    },
    {
      "key": "token",
      "value": ""
    }
  ]
}
```
