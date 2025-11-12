# Architecture Documentation

## System Overview

API Gateway built with Go using layered architecture for routing requests to backend microservices.

## Core Components

### 1. Entry Point (`cmd/gateway/main.go`)
- Application initialization
- Dependency injection
- Server lifecycle management

### 2. Middleware Stack
Execution order:
1. Recovery - Panic handling
2. Logging - Request/response logging
3. Request ID - Distributed tracing
4. CORS - Cross-origin support
5. Security Headers
6. Rate Limiter - Token bucket algorithm
7. JWT Auth - Token validation
8. Role Auth - Permission checking

### 3. Handlers
- **Auth Handler** - Registration, login, token refresh
- **Proxy Handler** - Request forwarding to services
- **Health Handler** - Liveness and readiness probes

### 4. Service Layer
- **Registry** - Service discovery and management
- **Load Balancer** - Round-robin distribution
- **Circuit Breaker** - Failure detection and recovery

### 5. Storage
- **MongoDB** - User data persistence
- **Redis** - Rate limiting and caching

## Request Flow

```
Client Request
  ↓
Middleware Stack
  ↓
Authentication/Authorization
  ↓
Service Registry
  ↓
Load Balancer
  ↓
Circuit Breaker
  ↓
Backend Service
```

## Security Layers

1. **Network** - HTTPS, firewall
2. **Gateway** - Rate limiting, validation
3. **Authentication** - JWT tokens
4. **Authorization** - Role-based access
5. **Data** - Encryption at rest/transit

## Design Patterns

- Gateway Pattern
- Middleware Pattern
- Registry Pattern
- Circuit Breaker Pattern
- Factory Pattern

## Scalability

- Stateless gateway design
- Horizontal scaling support
- Connection pooling
- Load balancing
