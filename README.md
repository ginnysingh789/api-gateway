# 🚀 API Gateway

> A production-ready API Gateway built with Go, featuring JWT authentication, rate limiting, load balancing, and circuit breakers for microservices architecture.


## ✨ Key Features

- 🔐 **JWT Authentication** - Secure token-based authentication with user management
- 🛡️ **Rate Limiting** - Token bucket algorithm preventing abuse (100 req/min per IP)
- ⚖️ **Load Balancing** - Round-robin distribution across service instances
- 🔌 **Circuit Breaker** - Automatic failure detection and recovery
- 📊 **Request Logging** - Structured JSON logs with request tracing
- 🏥 **Health Checks** - Kubernetes-ready liveness and readiness probes
- 🔒 **Security Headers** - HSTS, CSP, X-Frame-Options, and more
- 🌐 **CORS Support** - Configurable cross-origin resource sharing

## 📋 Prerequisites

- **Go** 1.21 or higher
- **Docker** & Docker Compose
- **MongoDB** 7.0+ (auto-started with Docker)
- **Redis** 7.0+ (auto-started with Docker)

## 🚀 Quick Start

### 1️⃣ Install Dependencies

```bash
cd Api-Gateway
go mod download && go mod tidy
```

### 2️⃣ Start with Docker (Recommended)

```bash
# Build and start all services (Gateway + MongoDB + Redis)
docker-compose -f deployments/docker-compose.yml build
docker-compose -f deployments/docker-compose.yml up -d

# Verify services are running
docker-compose -f deployments/docker-compose.yml ps
```

**Expected:** 3 containers running (gateway, mongo, redis)

### 3️⃣ Verify Gateway is Running

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": { "status": "healthy", "version": "1.0.0" }
}
```

---

## 🧪 API Testing

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

**Response:** Returns JWT token and user details

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}'
```

### Access Protected Route
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### PowerShell Testing Script

Save as `test-gateway.ps1` and run `.\test-gateway.ps1`:

```powershell
# Register User
$body = @{ username = "alice"; email = "alice@example.com"; password = "password123" } | ConvertTo-Json
$response = Invoke-WebRequest -Uri http://localhost:8080/api/v1/auth/register -Method POST -Body $body -ContentType "application/json"
$token = ($response.Content | ConvertFrom-Json).data.token

# Get Profile
$headers = @{ "Authorization" = "Bearer $token" }
Invoke-WebRequest -Uri http://localhost:8080/api/v1/profile -Method GET -Headers $headers
```

---

## 📡 API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/health` | Health check | No |
| GET | `/ready` | Readiness probe | No |
| POST | `/api/v1/auth/register` | Register new user | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/refresh` | Refresh JWT token | No |
| GET | `/api/v1/profile` | Get user profile | Yes |
| ANY | `/api/v1/users/*` | Proxy to users service | Yes |
| ANY | `/api/v1/products/*` | Proxy to products service | Yes |
| ANY | `/api/v1/orders/*` | Proxy to orders service | Yes |
| GET | `/api/v1/admin/services` | List services | Yes (Admin) |

📖 **Full API Documentation:** See [docs/API.md](docs/API.md)

---

## 🏗️ Project Structure

```
Api-Gateway/
├── cmd/gateway/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── middleware/      # Auth, rate limit, logging, CORS
│   ├── handler/         # HTTP request handlers
│   ├── service/         # Service registry & load balancer
│   ├── circuit/         # Circuit breaker implementation
│   └── models/          # Data models
├── pkg/
│   ├── logger/          # Structured logging
│   ├── storage/         # MongoDB & Redis clients
│   └── utils/           # JWT & response utilities
├── config/              # Configuration files
├── deployments/         # Docker & Kubernetes configs
└── docs/                # Documentation
```

---

## ⚙️ Configuration

Configuration via environment variables or `config/config.yaml`:

```yaml
server:
  port: 8080
  environment: production

jwt:
  secret: your-secret-key
  expiry: 24h

rate_limit:
  requests: 100
  window: 60s
```

**Environment Variables:**
```bash
PORT=8080
JWT_SECRET=your-secret-key
MONGO_URI=mongodb://localhost:27017
REDIS_ADDR=localhost:6379
```

---

## 🛑 Stopping Services

```bash
docker-compose -f deployments/docker-compose.yml down
```

---

## 🎯 Features in Detail

### Rate Limiting
- **Algorithm:** Token bucket with automatic refill
- **Default:** 100 requests per 60 seconds per IP
- **Storage:** Redis-backed for distributed rate limiting
- **Headers:** Returns `X-RateLimit-*` headers in responses

### Circuit Breaker
- **Threshold:** 5 consecutive failures
- **Timeout:** 30 seconds recovery period
- **States:** Closed → Open → Half-Open → Closed
- **Benefit:** Prevents cascading failures across services

### Security
- JWT token expiry: 24 hours (configurable)
- Password hashing: bcrypt with salt
- Security headers: HSTS, CSP, X-Frame-Options
- CORS: Configurable allowed origins

---

## 📚 Documentation

- **[API Documentation](docs/API.md)** - Complete API reference with examples
- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Configuration Guide](config/config.example.yaml)** - All configuration options

---

Made with ❤️ using Go

</div>
