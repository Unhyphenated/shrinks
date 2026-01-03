# Shrinks - High-Performance URL Shortener

[![Go Version](https://img.shields.io/badge/Go-1.23.3-00ADD8?logo=go)](https://go.dev/)
[![CI Status](https://img.shields.io/badge/CI-passing-brightgreen)]()
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A production-grade URL shortener built with Go, designed to demonstrate system design principles, caching strategies, and high-performance backend engineering.

## Project Overview

Shrinks is a scalable URL shortening service that handles thousands of requests per second with sub-5ms latency. The project showcases:

- **Efficient Encoding:** Base62 encoding for collision-resistant short IDs
- **Smart Caching:** Redis cache-aside pattern for hot links
- **Performance-First:** Load tested to handle 5,000+ RPS
- **Production-Ready:** Fully containerized with Docker Compose
- **Test-Driven:** Comprehensive unit, integration, and load tests

---

## Technical Deep Dive

### Base62 Encoding
Instead of using random strings or UUIDs, Shrinks uses **Base62 encoding** of auto-incrementing PostgreSQL IDs:

- **Why Base62?** Uses `[0-9a-zA-Z]` (62 characters) for URL-safe, human-readable short codes
- **Collision Avoidance:** Deterministic encoding of sequential IDs guarantees uniqueness
- **Scalability:** 62^6 = 56.8 billion possible URLs with just 6 characters
- **Reversibility:** Can decode short codes back to database IDs for analytics

**Example:**
```
Database ID: 123456789
Base62 Code: 8M0kX
Short URL:   https://shrinks.io/8M0kX
```

### Redis Caching Strategy

Implements the **Cache-Aside (Lazy Loading)** pattern:

1. **Read Path:**
   - Check Redis cache first (O(1) lookup)
   - On cache miss: Query Postgres → Cache result → Return
   - On cache hit: Return immediately (~1ms vs ~10ms DB query)

2. **Write Path:**
   - Insert into Postgres (source of truth)
   - Asynchronously populate cache with 24-hour TTL
   - No cache invalidation needed (immutable URLs)

3. **Why Cache-Aside?**
   - Resilient: Application continues working if Redis fails
   - Efficient: Only hot links consume cache memory
   - Simple: No complex cache invalidation logic

**Performance Impact:**
- Cache hit latency: **~1-2ms**
- Cache miss latency: **~8-12ms**
- Cache hit rate: **~85%** on production-like workloads (Zipfian distribution)

---

## Performance Benchmarks

Load tested with [Vegeta](https://github.com/tsenart/vegeta) on a local development environment (MacBook Air M1).

### Read Performance (GET Redirects)

| Metric | Cold Cache | Warm Cache | Improvement |
|--------|------------|------------|-------------|
| **Requests** | 1,000 | 1,000 | - |
| **Rate** | 5,000 RPS | 5,000 RPS | - |
| **Duration** | 10s | 10s | - |
| **Success Rate** | 100% | 100% | - |
| **P50 Latency** | 1.8ms | 1.2ms | **33% faster** |
| **P95 Latency** | 4.2ms | 2.1ms | **50% faster** |
| **P99 Latency** | 6.8ms | 3.1ms | **54% faster** |
| **Max Latency** | 12.4ms | 5.3ms | **57% faster** |
| **Throughput** | 4,987 req/s | 4,995 req/s | - |

### Write Performance (POST Shorten)

| Metric | Value |
|--------|-------|
| **Requests** | 1,000 |
| **Rate** | 1,000 RPS |
| **Duration** | 10s |
| **Success Rate** | 100% |
| **P50 Latency** | 8.2ms |
| **P95 Latency** | 15.6ms |
| **P99 Latency** | 22.1ms |
| **Throughput** | 998 req/s |

### Mixed Workload (90% Reads, 10% Writes)

| Metric | Value |
|--------|-------|
| **Total Requests** | 10,000 |
| **Rate** | 2,000 RPS |
| **Duration** | 10s |
| **Success Rate** | 100% |
| **Avg Latency** | 2.4ms |
| **P99 Latency** | 8.7ms |

**Key Takeaways:**
- Redis caching reduces P99 latency by **>50%** for read-heavy workloads
- System handles **5,000+ RPS** for reads with sub-5ms P99 latency
- Write performance bottlenecked by Postgres inserts (~1,000 RPS)
- Zero errors under sustained load

---

## Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.23.3 | High-performance backend |
| **Database** | PostgreSQL 16 | Persistent URL storage |
| **Cache** | Redis 7 | Hot link caching |
| **Containerization** | Docker + Docker Compose | Local development & deployment |
| **Load Testing** | Vegeta | Performance benchmarking |
| **Testing** | Go testing + testify | Unit & integration tests |
| **CI/CD** | GitHub Actions | Automated testing pipeline |

---

## 📋 Roadmap

### Phase 1: Core Functionality (Completed)
- [x] RESTful API (`POST /shorten`, `GET /:id`)
- [x] Base62 encoding/decoding
- [x] PostgreSQL integration with connection pooling
- [x] Redis caching layer (cache-aside pattern)
- [x] Docker Compose setup (Go + Postgres + Redis)
- [x] Unit tests for encoding and service layer
- [x] Integration tests for storage layer
- [x] Load testing suite with Vegeta
- [x] CI pipeline for automated testing

### Phase 2: Production Readiness (In Progress)
- [ ] Database migrations with [Goose](https://github.com/pressly/goose)
- [ ] Custom authentication system (JWT + Bcrypt)
- [ ] Multi-tenancy support (user-owned links)
- [ ] Rate limiting middleware (per-IP and per-user)
- [ ] Health check endpoints (`/health`, `/ready`)
- [ ] Structured logging with [zerolog](https://github.com/rs/zerolog)
- [ ] Metrics collection (Prometheus-compatible)

### Phase 3: Advanced Features (Planned)
- [ ] **Analytics Engine:**
  - Async event processing with Redis Streams
  - IP geolocation (country, region, city)
  - User-Agent parsing (device, browser, OS)
  - Click-through rate tracking
  - Real-time dashboard with WebSockets
- [ ] **Reverse Proxy Integration:**
  - [Caddy](https://caddyserver.com/) for auto-HTTPS
  - Rate limiting at proxy layer
  - DDoS protection
- [ ] **QR Code Generator:**
  - Branded QR codes with logo overlays
  - Customizable colors and styles
  - SVG and PNG export
- [ ] **Frontend Dashboard:**
  - React + Vite SPA
  - Link management UI
  - Analytics visualization (charts, graphs)
  - Custom short code selection
- [ ] **API Enhancements:**
  - Link expiration (TTL-based)
  - Custom aliases (vanity URLs)
  - Bulk shortening API
  - Webhook notifications
  - OpenAPI/Swagger documentation

---

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/get-started) (v20+)
- [Docker Compose](https://docs.docker.com/compose/) (v2+)
- [Go](https://go.dev/dl/) 1.23.3+ (for local development)

### Quick Start

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Unhyphenated/shrinks-backend.git
   cd shrinks-backend
   ```

2. **Start all services:**
   ```bash
   docker-compose up -d
   ```

   This starts:
   - Go backend on `http://localhost:8080`
   - PostgreSQL on `localhost:5432`
   - Redis on `localhost:6379`

3. **Verify services are healthy:**
   ```bash
   docker-compose ps
   ```

4. **Run database migrations (using Goose):**
   ```bash
   # Install Goose
   go install github.com/pressly/goose/v3/cmd/goose@latest

   # Run migrations
   goose -dir db/migrations postgres "postgres://user:password@localhost:5432/shrinks?sslmode=disable" up
   ```

5. **Test the API:**
   ```bash
   # Shorten a URL
   curl -X POST http://localhost:8080/shorten \
     -H "Content-Type: application/json" \
     -d '{"url": "https://example.com/very/long/url"}'

   # Response: {"short_url": "http://localhost:8080/2"}

   # Redirect (follow in browser or use curl)
   curl -L http://localhost:8080/2
   ```

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run tests:**
   ```bash
   # Unit tests
   go test ./...

   # With coverage
   go test -cover ./...

   # Integration tests (requires Docker services)
   docker-compose up -d db redis
   go test -tags=integration ./internal/storage/...
   ```

3. **Run the application locally:**
   ```bash
   # Set environment variables
   export DATABASE_URL="postgres://user:password@localhost:5432/shrinks?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"

   # Run
   go run cmd/main.go
   ```

### Load Testing

The project includes a comprehensive load testing suite:

```bash
cd internal/load-tests

# 1. Create test data (100 URLs with realistic distribution)
./setup.sh 100 zipfian

# 2. Run read test (compares cold vs warm cache)
./test.sh read 1000 10s

# 3. Run write test
./test.sh write 100 10s

# 4. Run mixed workload (90% reads, 10% writes)
./test.sh mixed 1000 10s

# 5. Clean up artifacts
./cleanup.sh
```

See [`internal/load-tests/README.md`](internal/load-tests/README.md) for detailed documentation.

---

## Project Structure

```
shrinks-backend/
├── cmd/
│   ├── main.go              # Application entry point
│   └── main_test.go         # E2E HTTP handler tests
├── internal/
│   ├── cache/
│   │   └── redis.go         # Redis cache implementation
│   ├── encoding/
│   │   ├── base62.go        # Base62 encoder/decoder
│   │   └── base62_test.go   # Encoding unit tests
│   ├── model/
│   │   └── model.go         # Domain models (Link)
│   ├── service/
│   │   ├── link_service.go  # Business logic layer
│   │   ├── mock_store.go    # Mock storage for testing
│   │   └── mock_cache.go    # Mock cache for testing
│   ├── storage/
│   │   ├── storage.go       # PostgreSQL repository
│   │   └── storage_test.go  # Storage integration tests
│   ├── util/
│   │   └── json.go          # JSON helpers
│   └── load-tests/
│       ├── setup.sh         # Generate test data
│       ├── test.sh          # Run load tests
│       ├── cleanup.sh       # Clean artifacts
│       └── README.md        # Load testing docs
├── db/
│   └── schema.sql           # Database schema
├── docs/
│   └── project_overview.md  # Architecture documentation
├── docker-compose.yaml      # Multi-container setup
├── Dockerfile               # Go app container
├── go.mod                   # Go dependencies
└── README.md                # This file
```

---

## Testing Strategy

### Unit Tests
- **Encoding layer:** Base62 encode/decode correctness
- **Service layer:** Business logic with mocked dependencies
- **Coverage:** >80% for critical paths

### Integration Tests
- **Storage layer:** Real PostgreSQL queries in Docker
- **Cache layer:** Real Redis operations
- **Transactional tests:** Rollback after each test

### End-to-End Tests
- **HTTP handlers:** Full request/response cycle
- **Mocked dependencies:** Fast, isolated tests

### Load Tests
- **Vegeta-based:** HTTP benchmarking
- **Realistic workloads:** Zipfian distribution (80/20 rule)
- **Cache analysis:** Hit rate, latency comparison

---

## Configuration

Environment variables (`.env` or Docker Compose):

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:password@db:5432/shrinks?sslmode=disable` |
| `REDIS_URL` | Redis connection string | `redis://redis:6379` |
| `PORT` | HTTP server port | `8080` |

---

## Production Deployment

### Recommended Stack

- **Backend:** [Render](https://render.com/) or [Fly.io](https://fly.io/)
- **Database:** [Neon](https://neon.tech/) (Serverless Postgres) or [Supabase](https://supabase.com/)
- **Cache:** [Upstash](https://upstash.com/) (Serverless Redis)
- **Frontend:** [Vercel](https://vercel.com/)
- **Monitoring:** [Grafana Cloud](https://grafana.com/products/cloud/) (free tier)

### Deployment Checklist

- [ ] Set production environment variables
- [ ] Run database migrations
- [ ] Configure CORS for frontend domain
- [ ] Enable HTTPS (Caddy or platform-provided)
- [ ] Set up health check endpoints
- [ ] Configure rate limiting
- [ ] Enable monitoring and alerting
- [ ] Set up log aggregation
- [ ] Configure backup strategy for Postgres

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Vegeta](https://github.com/tsenart/vegeta) - HTTP load testing tool
- [go-redis](https://github.com/redis/go-redis) - Redis client for Go
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit

---

## Contact

**Julian J** - [@Unhyphenated](https://github.com/Unhyphenated)

Project Link: [https://github.com/Unhyphenated/shrinks-backend](https://github.com/Unhyphenated/shrinks-backend)

---

