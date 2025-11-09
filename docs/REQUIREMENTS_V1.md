# create-go-service Requirements v1.0

## Overview

`create-go-service` generates production-ready Go microservices with complete CRUD operations for Posts and a CLI tool for service management.

## Scope

**Focus**: Chi REST and ConnectRPC gRPC frameworks with complete Posts CRUD example.

**Out of Scope (for now)**:
- API key management
- Rate limiting
- Huma framework

## TUI Configuration

### User Inputs
1. **Project Name**: e.g., "notes", "blog", "shop"
2. **Module Path**: e.g., `github.com/user/notes`
3. **CLI Name**: Auto-filled as `{project}ctl`, editable (e.g., `notectl`, `blogctl`)
4. **Output Directory**: Auto-filled from project name, editable
5. **API Framework** (single select):
   - Chi (REST)
   - ConnectRPC (gRPC)
6. **Database** (single select):
   - DynamoDB (preferred)
   - PostgreSQL/Supabase
7. **Features** (multi-select):
   - Metrics (Prometheus)
   - Authentication (JWT)
   - Hot Reload (wgo)

## Generated Service Structure

```
{project-name}/
├── cmd/
│   ├── api/                    # API server
│   │   └── main.go
│   └── {cli-name}/             # CLI tool (e.g., notectl)
│       └── main.go
├── internal/
│   ├── api/                    # API layer
│   │   ├── server.go
│   │   ├── routes.go           # Chi only
│   │   ├── handlers.go
│   │   └── json.go             # Chi only
│   ├── posts/                  # Posts CRUD
│   │   ├── post.go             # Domain model
│   │   ├── service.go          # Business logic
│   │   ├── table.go            # DynamoDB ops
│   │   ├── repository.go       # PostgreSQL ops
│   │   └── handlers.go         # HTTP/gRPC handlers
│   ├── database/
│   │   ├── dynamodb.go
│   │   └── postgres.go
│   ├── metrics/                # If selected
│   │   └── metrics.go
│   ├── auth/                   # If selected
│   │   ├── jwt.go
│   │   └── middleware.go
│   ├── config/
│   │   └── config.go
│   └── cli/                    # CLI commands
│       ├── root.go
│       ├── server.go
│       └── posts.go
├── protos/                     # gRPC only
│   ├── posts/v1/
│   │   └── posts.proto
│   └── gen/                    # Generated (gitignored)
├── config.yaml
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── fly.toml
├── wgo.yaml                    # If hot reload selected
├── Makefile
├── .github/workflows/
│   └── deploy.yml
├── go.mod
└── README.md
```

## Posts CRUD Implementation

### Domain Model

```go
type Post struct {
    ID        string    // Logical ID: "POST#<timestamp>"
    UserID    string    // Owner UUID
    Slug      string    // UUID for public URLs
    Title     string
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### DynamoDB Schema

```
Table: PostsTable

Primary Key:
- UserID (Partition Key): string
- ID (Sort Key): string - "POST#<timestamp>"

Attributes:
- Slug: string (UUID)
- Title: string
- Content: string
- CreatedAt: timestamp (ISO8601)
- UpdatedAt: timestamp (ISO8601)

GSI: SlugIndex
- Slug (Partition Key): UUID
- ProjectionType: ALL
```

### PostgreSQL Schema

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
CREATE UNIQUE INDEX idx_posts_slug ON posts(slug);
```

### API Endpoints (Chi)

```
POST   /api/v1/posts              # Create post
GET    /api/v1/posts              # List user's posts
GET    /api/v1/posts/:slug        # Get post by slug
PUT    /api/v1/posts/:slug        # Update post
DELETE /api/v1/posts/:slug        # Delete post
```

### gRPC Service (ConnectRPC)

```protobuf
service PostsService {
  rpc CreatePost(CreatePostRequest) returns (Post) {}
  rpc GetPost(GetPostRequest) returns (Post) {}
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse) {}
  rpc UpdatePost(UpdatePostRequest) returns (Post) {}
  rpc DeletePost(DeletePostRequest) returns (google.protobuf.Empty) {}
}
```

## CLI Tool

### Name
User-selectable, default: `{project}ctl`
Examples: `notectl`, `blogctl`, `shopctl`, `apictl`

### Commands

```bash
# Server management
{cli} server start              # Start API server
{cli} server seed               # Seed test data

# Posts CRUD
{cli} posts create --title "..." --content "..." --user-id <uuid>
{cli} posts list --user-id <uuid>
{cli} posts get --slug <uuid>
{cli} posts update --slug <uuid> --title "..." --content "..."
{cli} posts delete --slug <uuid>

# Version
{cli} version
```

### Features
- Cobra framework
- Colored output (lipgloss)
- JSON/table output formats
- Shell completion
- Config file support (~/.{cli}/config.yaml)

## Configuration

### YAML (config.yaml)
```yaml
server:
  port: "8080"
  environment: "local"
  enable_reflection: true  # gRPC only

database:
  # DynamoDB
  aws_region: "us-east-1"
  table_name: "{project}"
  endpoint_url: "http://localhost:8000"
  
  # PostgreSQL
  # host: "localhost"
  # port: 5432
  # database: "{project}"

metrics:  # If enabled
  enabled: true
  path: "/metrics"
```

### Environment Variables (.env)
```bash
# Database
DATABASE_URL=postgres://...      # PostgreSQL
DYNAMODB_ENDPOINT=http://...     # DynamoDB local

# Auth (if enabled)
JWT_SECRET=your-secret-key

# Deployment
FLY_API_TOKEN=...
```

## Dependencies

### Core
- `github.com/google/uuid`
- `github.com/caarlos0/env/v10`
- `gopkg.in/yaml.v3`

### Chi
- `github.com/go-chi/chi/v5`

### ConnectRPC
- `connectrpc.com/connect`
- `connectrpc.com/grpcreflect`
- `golang.org/x/net`
- `google.golang.org/protobuf`

### Database
- **DynamoDB**: `github.com/aws/aws-sdk-go-v2/*`
- **PostgreSQL**: `github.com/jackc/pgx/v5`, `github.com/georgysavva/scany/v2`

### Optional Features
- **Auth**: `github.com/golang-jwt/jwt/v5`
- **Metrics**: `github.com/prometheus/client_golang`

### CLI
- `github.com/spf13/cobra`
- `github.com/charmbracelet/lipgloss`

## Ready-to-Run Features

✅ **Complete Posts CRUD**
- Create, Read, Update, Delete operations
- Slug-based lookups
- User-scoped queries

✅ **CLI Tool**
- All CRUD operations via CLI
- Server management
- Seed data command

✅ **Database Setup**
- Automatic table creation
- Local development (DynamoDB Local / PostgreSQL)
- Production-ready schemas

✅ **Deployment**
- Fly.io configuration
- GitHub Actions CI/CD
- Docker support

✅ **Development**
- Hot reload (optional)
- Docker Compose
- Example data

✅ **Documentation**
- README with examples
- .env.example
- API documentation

## Development Workflow

```bash
# Generate service
create-go-service

# Setup
cd my-service
cp .env.example .env
# Edit .env

# Start dependencies
docker compose up -d

# Run server
make run

# Use CLI
make cli-install
{cli} posts create --title "Test" --user-id $(uuidgen)

# Deploy
make deploy
```

## Success Criteria

Generated services must:
1. ✅ Compile without errors
2. ✅ Run locally with docker-compose
3. ✅ Perform all CRUD operations
4. ✅ CLI tool works for all operations
5. ✅ Deploy to Fly.io successfully
6. ✅ Include working examples
7. ✅ Have complete documentation

## Implementation Priority

### Phase 1 (Current Focus)
1. ✅ Chi REST API with Posts CRUD
2. ✅ ConnectRPC gRPC with Posts CRUD
3. ✅ DynamoDB integration
4. ✅ PostgreSQL integration
5. ✅ CLI tool (Cobra-based)
6. ✅ Configuration system (YAML + env)
7. ✅ Docker & docker-compose
8. ✅ Fly.io deployment

### Phase 2 (Future)
- API key management
- Rate limiting (Upstash Redis)
- Huma framework
- Temporal workflows
- Message queues
- Additional deployment targets

