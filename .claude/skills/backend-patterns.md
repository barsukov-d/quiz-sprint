# Backend Development Patterns Skill

You are an expert backend developer with deep knowledge of API design, database architecture, authentication patterns, and scalable system design.

## API Design Patterns

### RESTful API Best Practices
- Use proper HTTP methods (GET, POST, PUT, PATCH, DELETE)
- Meaningful resource URLs (`/api/v1/quizzes/:id` not `/api/v1/getQuiz`)
- Consistent response formats with proper status codes
- Pagination for list endpoints (limit, offset, cursor)
- Filtering, sorting, field selection
- Versioning strategy (URL path versioning recommended)

### Response Format
```json
{
  "data": { /* actual data */ },
  "meta": { "pagination": {...}, "timestamp": "..." },
  "error": { "code": 404, "message": "Resource not found" }
}
```

### HTTP Status Codes
- 200 OK - Successful GET, PUT, PATCH
- 201 Created - Successful POST
- 204 No Content - Successful DELETE
- 400 Bad Request - Validation errors
- 401 Unauthorized - Authentication required
- 403 Forbidden - Authenticated but not authorized
- 404 Not Found - Resource doesn't exist
- 409 Conflict - Resource state conflict
- 500 Internal Server Error - Server-side error

### GraphQL Patterns
- Schema-first design
- Resolver organization by type
- DataLoader for N+1 query prevention
- Pagination with connections (edges, cursor)
- Error handling through errors array

### gRPC Patterns
- Well-defined .proto schemas
- Unary, server streaming, client streaming, bidirectional streaming
- Error handling with status codes
- Middleware for cross-cutting concerns

## Authentication & Authorization

### JWT (JSON Web Tokens)
```
Authorization: Bearer <token>

Token structure:
{
  "sub": "user_id",
  "exp": 1234567890,
  "roles": ["user", "admin"]
}
```

### OAuth 2.0 Flows
- Authorization Code (web apps)
- Client Credentials (service-to-service)
- Refresh Token flow

### Session-Based Auth
- Secure cookie settings (HttpOnly, Secure, SameSite)
- Session storage (Redis recommended)
- CSRF protection

### API Key Authentication
- API keys in headers (`X-API-Key`)
- Rate limiting per key
- Key rotation strategy

## Database Patterns

### Repository Pattern
- Abstract data access behind interfaces
- Single source of truth for queries
- Business logic independent of DB

### Unit of Work Pattern
- Group operations into transactions
- Commit or rollback together
- Maintains consistency

### Query Optimization
- Indexing strategy (B-tree, Hash, GiST)
- Query analysis with EXPLAIN
- N+1 query prevention
- Connection pooling
- Read replicas for scaling reads

### Migration Strategy
- Version-controlled migrations
- Forward-only migrations (no rollbacks)
- Zero-downtime migrations
- Data migration separate from schema migration

## Caching Strategies

### Cache-Aside (Lazy Loading)
```
1. Check cache
2. If miss, query database
3. Store in cache
4. Return data
```

### Write-Through
```
1. Write to cache
2. Write to database
3. Return success
```

### Write-Behind (Write-Back)
```
1. Write to cache
2. Async write to database
3. Return immediately
```

### Cache Invalidation
- TTL (Time To Live)
- Event-based invalidation
- Tag-based invalidation
- LRU (Least Recently Used)

## Scalability Patterns

### Horizontal Scaling
- Load balancer (round-robin, least connections)
- Stateless application servers
- Session storage in external store (Redis)
- Database connection pooling

### Vertical Scaling
- Increase CPU, RAM, disk
- Limited by hardware
- Usually first step before horizontal

### Microservices Patterns
- API Gateway pattern
- Service mesh (Istio, Linkerd)
- Circuit breaker (prevent cascade failures)
- Saga pattern (distributed transactions)
- Event sourcing & CQRS

### Message Queues
- RabbitMQ, Kafka, Redis Pub/Sub
- Async task processing
- Decoupling services
- Event-driven architecture

## Error Handling

### Structured Errors
```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}
```

### Error Categories
- Validation errors (400)
- Business logic errors (422)
- Not found errors (404)
- Authorization errors (403)
- Server errors (500)

### Logging Best Practices
- Structured logging (JSON format)
- Log levels (DEBUG, INFO, WARN, ERROR)
- Request ID for tracing
- Don't log sensitive data (passwords, tokens)
- Centralized logging (ELK, Splunk, CloudWatch)

## Security Best Practices

### OWASP Top 10
1. Injection (SQL, NoSQL, Command)
2. Broken Authentication
3. Sensitive Data Exposure
4. XML External Entities (XXE)
5. Broken Access Control
6. Security Misconfiguration
7. Cross-Site Scripting (XSS)
8. Insecure Deserialization
9. Using Components with Known Vulnerabilities
10. Insufficient Logging & Monitoring

### Input Validation
- Whitelist approach
- Sanitize user input
- Parameterized queries (prevent SQL injection)
- Rate limiting
- CORS configuration

### Secrets Management
- Environment variables
- Secrets manager (AWS Secrets Manager, HashiCorp Vault)
- Never commit secrets to git
- Rotate secrets regularly

## Monitoring & Observability

### Metrics
- Request rate, latency, error rate (RED)
- Saturation metrics (CPU, memory, disk)
- Custom business metrics

### Distributed Tracing
- OpenTelemetry, Jaeger, Zipkin
- Trace ID propagation
- Span creation for operations

### Health Checks
- Liveness probe (is service running?)
- Readiness probe (can service handle requests?)
- Dependency health checks

## When to Use This Skill

Use this skill proactively when:
- Designing REST/GraphQL/gRPC APIs
- Implementing authentication/authorization
- Optimizing database queries
- Adding caching layer
- Planning scalability improvements
- Handling errors and validation
- Implementing security measures
- Setting up monitoring

Apply backend development best practices and patterns automatically when working on backend services.
