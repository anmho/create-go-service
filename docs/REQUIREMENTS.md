# create-go-service Requirements

## Overview

`create-go-service` generates production-ready Go microservices with complete CRUD operations, API key management, rate limiting, and a CLI tool.

## Core Requirements

### 1. Project Configuration
- **Project Name**: User-provided (e.g., "notes", "blog", "shop")
- **Module Path**: Go module path (e.g., github.com/user/notes)
- **CLI Name**: Auto-filled as `{project}ctl`, editable (e.g., notectl, blogctl, shopctl)
- **Output Directory**: Auto-filled from project name, editable

### 2. API Frameworks (Single Select)
- **Chi** (REST): Lightweight, idiomatic Go router
- **ConnectRPC** (gRPC): Modern gRPC with HTTP/2

### 3. Database (Single Select)
- **DynamoDB** (Preferred): Serverless, scalable NoSQL
- **PostgreSQL/Supabase**: Relational database with pgx driver

### 4. Features (Multi-Select)
- **Metrics**: Prometheus instrumentation
- **Authentication**: JWT-based auth
- **API Keys & Rate Limiting**: Upstash Redis-based rate limiting
- **Hot Reload**: wgo for development

## Generated Service Features

### Posts CRUD (Example Feature)

**DynamoDB Schema**:
```
Table: PostsTable
Primary Key: (UserID, PostID)
- UserID (Partition Key): string - User's UUID
- PostID (Sort Key): string - "POST#<timestamp>"

Attributes:
- Slug: string (UUID) - Public identifier for URLs
- Title: string
- Content: string
- CreatedAt: timestamp (ISO8601)
- UpdatedAt: timestamp (ISO8601)

GSI: SlugIndex
- Slug (Partition Key): UUID
- ProjectionType: ALL
```

**Operations**:
```
POST   /api/v1/posts              # Create post
GET    /api/v1/posts              # List user's posts
GET    /api/v1/posts/:slug        # Get post by slug
PUT    /api/v1/posts/:slug        # Update post
DELETE /api/v1/posts/:slug        # Delete post
```

**PostgreSQL Schema**:
```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    slug UUID NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_slug ON posts(slug);
```

### API Key Management

**Schema**:
```
DynamoDB:
Primary Key: (UserID, KeyID)
- UserID (Partition Key): Owner's UUID
- KeyID (Sort Key): "KEY#<uuid>"

Attributes:
- Key: string (bcrypt hashed)
- Name: string
- Prefix: string (first 8 chars, e.g., "sk_live_")
- Scopes: []string (e.g., ["posts:read", "posts:write"])
- RateLimit: int (requests per minute)
- CreatedAt: timestamp
- LastUsedAt: timestamp
- ExpiresAt: timestamp (optional)

GSI: KeyPrefixIndex
- Prefix (Partition Key)
```

**Key Format**:
```
sk_live_<32_random_chars>  # Production
sk_test_<32_random_chars>  # Testing
```

**Operations**:
```
POST   /api/v1/keys               # Generate new key
GET    /api/v1/keys               # List user's keys
DELETE /api/v1/keys/:keyId        # Revoke key
```

### Rate Limiting

**Implementation**: Upstash Redis (serverless, HTTP-based)

**Features**:
- Per-API-key rate limits
- Sliding window algorithm
- Rate limit headers (X-RateLimit-*)
- Default: 100 requests/minute
- Configurable per key

**Middleware**:
```go
// Automatic rate limiting on all API routes
X-API-Key header → validate → check rate limit → allow/deny
```

**Configuration**:
```yaml
# config.yaml
ratelimit:
  enabled: true
  default_limit: 100  # req/min
  burst: 20

# .env
UPSTASH_REDIS_URL=https://your-redis.upstash.io
UPSTASH_REDIS_TOKEN=your-token
```

### CLI Tool

**Name**: `{project}ctl` (e.g., `notectl`, `blogctl`, `shopctl`)

**Commands**:
```bash
# Server management
{cli} server start          # Start API server
{cli} server migrate        # Run migrations
{cli} server seed           # Seed test data

# Posts CRUD
{cli} posts create --title "..." --content "..." --user-id <uuid>
{cli} posts list --user-id <uuid>
{cli} posts get --slug <uuid>
{cli} posts update --slug <uuid> --title "..."
{cli} posts delete --slug <uuid>

# API key management
{cli} keys create --name "..." --scopes "..." --rate-limit 1000
{cli} keys list --user-id <uuid>
{cli} keys revoke --key-id <uuid>

# Version
{cli} version
```

**Features**:
- Cobra framework
- Colored output (lipgloss)
- JSON/table output formats
- Shell completion (bash/zsh/fish)
- Config file support (~/.{cli}/config.yaml)

## Project Structure

```
generated-service/
├── cmd/
│   ├── api/                    # API server
│   │   └── main.go
│   └── {cli-name}/             # CLI tool
│       └── main.go
├── internal/
│   ├── api/                    # API layer
│   │   ├── server.go
│   │   ├── routes.go           # Chi routes
│   │   ├── handlers.go
│   │   └── json.go
│   ├── posts/                  # Posts feature
│   │   ├── post.go             # Model
│   │   ├── service.go          # Business logic
│   │   ├── table.go            # DynamoDB ops
│   │   ├── repository.go       # PostgreSQL ops
│   │   └── handlers.go         # HTTP handlers
│   ├── apikeys/                # API key management
│   │   ├── apikey.go
│   │   ├── service.go
│   │   ├── table.go
│   │   └── handlers.go
│   ├── ratelimit/              # Rate limiting
│   │   ├── middleware.go
│   │   └── redis.go            # Upstash client
│   ├── auth/                   # Authentication
│   │   ├── jwt.go
│   │   └── middleware.go
│   ├── database/               # DB clients
│   │   ├── dynamodb.go
│   │   └── postgres.go
│   ├── metrics/                # Prometheus
│   │   └── metrics.go
│   ├── config/                 # Configuration
│   │   └── config.go
│   └── cli/                    # CLI commands
│       ├── root.go
│       ├── server.go
│       ├── posts.go
│       └── keys.go
├── protos/                     # gRPC only
│   ├── posts/v1/
│   │   └── posts.proto
│   └── gen/                    # Generated (gitignored)
├── config.yaml                 # Non-sensitive config
├── .env.example                # Env template
├── Dockerfile
├── docker-compose.yml
├── fly.toml
├── wgo.yaml
├── Makefile
├── .github/workflows/
│   └── deploy.yml
├── go.mod
└── README.md
```

## Dependencies

### Core
- `github.com/google/uuid` - UUID generation
- `github.com/caarlos0/env/v10` - Env var loading
- `gopkg.in/yaml.v3` - YAML parsing

### Chi (REST)
- `github.com/go-chi/chi/v5` - Router

### ConnectRPC (gRPC)
- `connectrpc.com/connect` - Connect protocol
- `connectrpc.com/grpcreflect` - Reflection
- `golang.org/x/net` - HTTP/2
- `google.golang.org/protobuf` - Protobuf

### Database
- **DynamoDB**: `github.com/aws/aws-sdk-go-v2/*`
- **PostgreSQL**: `github.com/jackc/pgx/v5`, `github.com/georgysavva/scany/v2`

### Features
- **Auth**: `github.com/golang-jwt/jwt/v5`
- **Metrics**: `github.com/prometheus/client_golang`
- **Rate Limiting**: `github.com/upstash/upstash-redis-go` (HTTP-based)
- **CLI**: `github.com/spf13/cobra`, `github.com/charmbracelet/lipgloss`
- **Security**: `golang.org/x/crypto/bcrypt` (API key hashing)

## Configuration

### YAML (config.yaml)
```yaml
server:
  port: "8080"
  environment: "local"

database:
  # DynamoDB or PostgreSQL config

ratelimit:
  enabled: true
  default_limit: 100

metrics:
  enabled: true
  path: "/metrics"
```

### Environment Variables (.env)
```bash
# Database
DATABASE_URL=postgres://...      # PostgreSQL
DYNAMODB_ENDPOINT=http://...     # DynamoDB local

# Rate Limiting
UPSTASH_REDIS_URL=https://...
UPSTASH_REDIS_TOKEN=...

# Auth
JWT_SECRET=...

# Deployment
FLY_API_TOKEN=...
```

## Deployment

### Fly.io
- Single instance deployment
- Automatic scaling
- Secrets via `flyctl secrets`
- GitHub Actions CI/CD

### Docker
- Multi-stage build
- Local development with docker-compose
- DynamoDB Local or PostgreSQL containers

## Development Workflow

```bash
# Generate service
create-go-service

# Setup
cd my-service
cp .env.example .env
# Edit .env with secrets

# Start dependencies
docker compose up -d

# Run server
make run

# Or with hot reload
wgo run cmd/api/main.go

# Use CLI
make cli-install
{cli} posts create --title "Test" --user-id <uuid>

# Deploy
make deploy
```

## Ready-to-Run

Generated services are **immediately functional** with:
- ✅ Complete CRUD operations for Posts
- ✅ API key generation and validation
- ✅ Rate limiting with Upstash
- ✅ CLI tool for all operations
- ✅ Database setup (local & production)
- ✅ Deployment configs
- ✅ CI/CD pipeline
- ✅ Documentation
- ✅ Example requests

No additional setup required - just add your business logic!

