# create-go-service Design Document

## Executive Summary

`create-go-service` is a CLI tool that scaffolds production-ready Go microservice boilerplate with a modern TUI interface. It generates services with folder-by-feature architecture, supports multiple API styles (REST with Chi/Huma, gRPC with ConnectRPC), includes metrics instrumentation, deployment configurations for Fly.io, and development tooling like wgo for hot reload.

The tool uses this repository (`create-go-service`) as the reference template and inspiration for project structure, demonstrating best practices for Go microservice development.

## 1. Architecture Overview

### 1.1 CLI Tool Structure

```
create-go-service/
├── cmd/
│   └── create-go-service/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cli/                      # CLI command handlers
│   │   ├── create.go            # Main create command
│   │   └── deploy.go            # Deploy command (future)
│   ├── tui/                      # Bubbletea TUI components
│   │   ├── main.go              # Main TUI model
│   │   ├── forms.go             # Form components
│   │   ├── menus.go             # Menu navigation
│   │   └── progress.go          # Progress indicators
│   ├── generator/                # Code generation logic
│   │   ├── project.go           # Project structure generator
│   │   ├── files.go             # File generation
│   │   ├── templates.go         # Template loading (embedded)
│   │   ├── dependencies.go      # go.mod management
│   │   └── templates/           # Embedded template files (go:embed)
│   │       ├── base/            # Common files
│   │       ├── chi/             # Chi REST templates
│   │       ├── huma/            # Huma REST templates
│   │       ├── grpc/            # ConnectRPC templates
│   │       ├── metrics/         # Prometheus metrics
│   │       ├── config/          # Config structs
│   │       ├── docker/          # Dockerfile, docker-compose
│   │       ├── fly/             # Fly.io configs
│   │       └── [other features]
│   └── config/                   # CLI tool config
│       └── config.go            # User preferences
```

### 1.2 Generated Service Structure

The generated service follows folder-by-feature pattern (inspired by this repo):

```
generated-service/
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entry point
│   └── postctl/                 # CLI tool
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go            # Config struct with YAML + env loading
│   ├── api/                     # API layer (REST/gRPC)
│   │   ├── server.go            # Server setup
│   │   ├── routes.go            # Route registration (Chi)
│   │   ├── handlers.go          # HTTP handlers (Chi)
│   │   ├── json.go              # JSON helpers (Chi)
│   │   └── middleware.go        # Middleware (auth, metrics, etc.)
│   ├── posts/                   # Example feature: Posts CRUD
│   │   ├── post.go              # Domain model
│   │   ├── service.go           # Business logic
│   │   ├── table.go             # DynamoDB operations (if DynamoDB)
│   │   ├── repository.go        # PostgreSQL operations (if PostgreSQL)
│   │   └── handlers.go          # HTTP/gRPC handlers
│   ├── database/                # Database clients
│   │   ├── dynamodb.go          # DynamoDB client setup
│   │   └── postgres.go          # PostgreSQL client setup
│   ├── metrics/                 # Metrics instrumentation (optional)
│   │   └── metrics.go           # Prometheus metrics
│   ├── auth/                    # Authentication (optional)
│   │   ├── jwt.go               # JWT/auth logic
│   │   └── middleware.go        # Auth middleware
│   └── cli/                     # CLI commands
│       ├── root.go              # Root command
│       ├── posts.go             # Posts commands
│       └── server.go            # Server management commands
├── protos/                       # Protocol buffer definitions (gRPC only)
│   ├── posts/
│   │   └── v1/
│   │       └── posts.proto      # Posts service definition
│   └── gen/                     # Generated code (gitignored)
├── Dockerfile
├── docker-compose.yml
├── config.yaml                   # Non-sensitive configuration
├── .env.example                  # Example environment variables
├── fly.toml                      # Fly.io deployment config
├── wgo.yaml                      # wgo hot reload config
├── Makefile                      # Makefile with deploy command
├── .github/
│   └── workflows/
│       └── deploy.yml            # GitHub Actions deployment
├── go.mod
└── README.md
```

**Generated Services Include**:

- Complete CRUD operations for Posts
- Proper DynamoDB schema with UserID + PostID composite key
- Slug-based lookups using GSI
- CLI tool for service management (e.g., `notectl`)
- Request/response validation
- Error handling
- Ready-to-run examples
- Optional: JWT authentication
- Optional: Prometheus metrics

## 2. TUI Design (Bubbletea)

### 2.1 User Flow

1. **Welcome Screen**: Tool introduction and version
2. **Project Configuration**:
   - Project name
   - Module path (e.g., github.com/user/service)
   - Output directory
3. **API Framework Selection** (single-select):
   - REST with Chi
   - REST with Huma (includes Swagger)
   - gRPC with ConnectRPC
4. **Database Selection**:
   - DynamoDB (default, preferred)
   - Supabase/PostgreSQL (with pgx)
   - Option to serialize protos (DynamoDB only)
5. **Features Selection**:
   - Metrics (Prometheus/GMP)
   - Authentication (JWT)
   - Hot reload (wgo)
6. **Deployment Selection**:
   - Fly.io (default)
7. **Review & Generate**: Show summary, confirm generation

### 2.2 TUI Components

- **Form inputs**: Text input with styled prompts and cursors
- **Selection components**:
  - Single-select radio buttons for API framework, database, deployment
  - Multi-select checkboxes for features
- **Navigation**: Arrow keys (↑/↓), j/k, Space, Enter, Esc
- **Animated spinners**: Dot spinner during generation with step-by-step progress
- **Progress indicators**:
  - Real-time generation steps with checkmarks
  - Current step highlighted with spinner
  - Pending steps shown in gray
- **Status messages**: Success/error with styled icons and colors
- **Color scheme**:
  - Primary blue (#39) for highlights
  - Secondary green (#86) for accents
  - Success green (#46) for completed items
  - Error red (#196) for failures
  - Gray (#240) for secondary text

## 3. Template System

### 3.1 Template Organization

Templates are **embedded into the executable** using Go's `embed` package (`//go:embed` directive), making the binary self-contained and portable. No external template files are needed at runtime.

Templates organized by feature and API type:

```
internal/generator/templates/     # Embedded in binary via //go:embed
├── base/                         # Common files
│   ├── go.mod.tmpl
│   ├── README.md.tmpl
│   ├── .gitignore.tmpl
│   └── Makefile.tmpl
├── chi/                          # Chi REST templates
│   ├── internal/api/server.go.tmpl
│   ├── internal/api/routes.go.tmpl
│   └── internal/api/handlers.go.tmpl
├── huma/                         # Huma REST templates
│   ├── internal/api/server.go.tmpl
│   ├── internal/api/routes.go.tmpl
│   └── swagger.yaml.tmpl
├── grpc/                         # ConnectRPC templates
│   ├── protos/service.proto.tmpl
│   ├── internal/api/server.go.tmpl
│   └── internal/api/handlers.go.tmpl
├── dynamodb/                     # DynamoDB templates
│   ├── internal/database/dynamodb.go.tmpl
│   └── internal/[feature]/[feature]_table.go.tmpl
├── postgres/                     # PostgreSQL/Supabase templates
│   ├── internal/database/postgres.go.tmpl
│   ├── internal/[feature]/[feature]_repository.go.tmpl
│   └── migrations/               # Migration templates (optional)
│       └── 001_initial.up.sql.tmpl
├── metrics/                      # Metrics templates
│   └── internal/metrics/metrics.go.tmpl
├── config/                       # Config templates
│   └── internal/config/config.go.tmpl
├── docker/                       # Docker templates
│   ├── Dockerfile.tmpl
│   └── docker-compose.yml.tmpl
├── fly/                          # Fly.io templates
│   └── fly.toml.tmpl
├── github/                       # GitHub Actions templates
│   └── workflows/
│       └── deploy.yml.tmpl
└── makefile/                     # Makefile templates
    └── Makefile.tmpl
```

### 3.2 Template Variables

Common template variables:

- `{{.ProjectName}}`: Project name
- `{{.ModulePath}}`: Go module path
- `{{.APIType}}`: Selected API type (chi, huma, grpc)
- `{{.Database}}`: Database type (dynamodb, postgres)
- `{{.HasMetrics}}`: Boolean for metrics inclusion
- `{{.HasAuth}}`: Boolean for auth inclusion
- `{{.Features}}`: List of feature modules

## 4. API Options

### 4.1 REST with Chi

**Structure** (based on current repo):

- Chi router with middleware
- Centralized error handling
- Prometheus-style dependency injection with currying
- Example: `internal/api/routes.go` pattern

**Key Features**:

- Middleware chain (logging, recovery, metrics)
- Handler functions with service injection
- JSON response helpers
- Error response standardization

### 4.2 REST with Huma

**Structure** (inspired by happened repo):

- Huma framework with OpenAPI/Swagger
- Automatic API documentation
- Request/response validation
- Similar folder structure to Chi option

**Key Features**:

- OpenAPI 3.0 spec generation
- Swagger UI endpoint
- Type-safe request/response handling
- Built-in validation

### 4.3 gRPC with ConnectRPC

**Structure**:

- Protocol buffer definitions in `protos/` directory
- ConnectRPC server setup
- RPC handlers by feature
- Remote publishing setup included

**Key Features**:

- `.proto` file generation in `protos/` directory
- ConnectRPC server configuration
- gRPC-Web support
- **Reflection service**: Enabled automatically in local stage, disabled in production
  - Stage-based: enabled when `STAGE=local`, disabled otherwise
  - No explicit flag needed - purely stage-based
  - Used for tools like `grpcurl`, `buf`, etc.

## 5. Database Integration

### 5.1 DynamoDB (Preferred)

**Structure** (based on current repo):

- DynamoDB client in `internal/database/`
- Feature-specific table operations (e.g., `posts/table.go`)
- Table creation helpers
- Local DynamoDB support

**Schema Design - Posts Example**:

```
Table: PostsTable

Primary Key:
- UserID (Partition Key): string - User's UUID
- PostID (Sort Key): string - Logical post ID (e.g., "POST#<timestamp>")

Attributes:
- Slug: string (UUID) - Public identifier for URL routing
- Title: string
- Content: string
- CreatedAt: timestamp (ISO8601)
- UpdatedAt: timestamp (ISO8601)

Global Secondary Index: SlugIndex
- Slug (Partition Key): UUID string
- Allows O(1) lookup by slug for public access
- ProjectionType: ALL
```

**CRUD Operations**:

```go
// Create
Put(UserID, PostID, {Slug, Title, Content, ...})

// Read by slug (public access)
Query(SlugIndex, Slug = "uuid-here")

// Read user's posts
Query(PostsTable, UserID = "user-uuid")

// Read single post
GetItem(UserID, PostID)

// Update
UpdateItem(UserID, PostID, {Title, Content, UpdatedAt})

// Delete
DeleteItem(UserID, PostID)
```

**Key Features**:

- AWS SDK v2 integration
- Composite key design (UserID + PostID)
- GSI for slug-based lookups
- Query/scan operations
- Local development setup
- Automatic table creation

### 5.2 Supabase/PostgreSQL

**Structure**:

- PostgreSQL client using `pgx` driver
- Feature-specific repository operations
- Migration support (optional)
- Connection pooling

**Key Features**:

- **pgx** driver with native struct scanning:
  - Built-in `pgx.RowToStructByName` and `pgx.CollectRows` for automatic struct scanning
  - Type-safe query results
  - Native scanning without external dependencies
  - Batch operations support
- **Atlas Go migrations**:
  - Migration files created via Atlas CLI
  - Up/down migration support
  - Reproducible migrations
  - Applied via `make deploy` or `make migrate-prod`
- Provider-agnostic:
  - Works with any PostgreSQL provider (Supabase, AWS RDS, etc.)
  - Connection string format
  - SSL/TLS configuration
  - Connection pooling
- Local development:
  - Docker Compose with PostgreSQL
  - Testcontainers for testing
  - Migration scripts

**Alternative SQL Libraries Considered**:

- **pgx** (chosen): Modern, performant, native struct scanning (no scany needed)
- **sqlx**: Good but less performant than pgx
- **sqlc**: Code generation, but adds complexity
- **gorm**: ORM approach, but may be overkill for microservices

**Implementation**:

```go
// Example with pgx native scanning
import "github.com/jackc/pgx/v5"

type User struct {
    ID    uuid.UUID
    Name  string
    Email string
}

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    rows, _ := r.pool.Query(ctx, "SELECT id, name, email FROM users WHERE id = $1", id)
    user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
    return &user, err
}
```

## 6. PostHog Integration (Optional)

### 6.1 PostHog Analytics

**Structure**:

- PostHog client in `internal/posthog/`
- Interface-based design for easy mocking
- Event tracking for user actions
- Startup validation to ensure client is never nil when enabled

**Key Features**:

- **Client Interface**: `posthog.Client` interface for easy testing
- **Event Tracking**: Capture user events (post_created, post_viewed, etc.)
- **User Identification**: Identify users for analytics
- **Optional**: Only included when PostHog feature is selected
- **Startup Validation**: Panics if PostHog is enabled but client is nil

**Implementation**:

```go
// PostHog client interface
type Client interface {
    Capture(ctx context.Context, distinctID string, event string, properties map[string]interface{}) error
    Identify(ctx context.Context, distinctID string, properties map[string]interface{}) error
    Close() error
}

// Usage in handlers
posthogClient.Capture(ctx, userID.String(), "post_created", map[string]interface{}{
    "post_id": post.ID.String(),
    "title":   post.Title,
})
```

## 7. Metrics Integration (Always Included)

### 7.1 Prometheus Metrics

**Structure**:

- Metrics package in `internal/metrics/`
- Standard HTTP metrics (request duration, count)
- Custom business metrics
- Metrics endpoint (`/metrics`)

**Key Features**:

- Request duration histogram
- Request counter
- Error counter
- Custom business metrics
- Compatible with GMP (Google Managed Prometheus) and self-hosted Prometheus

## 8. CLI Tool

### 8.1 Generated CLI Structure

Each generated service includes a CLI tool (using Cobra) for service management and operations.

**CLI Name**: User-selectable during generation

- Default pattern: `{project}ctl` (e.g., `notectl`, `blogctl`, `apictl`)
- Examples: `notectl`, `appctl`, `gitctl`, `shopctl`

**Commands Structure**:

```
postctl
├── server              # Server management
│   ├── start          # Start the API server
│   └── seed           # Seed test data
├── posts              # Posts CRUD operations
│   ├── create         # Create a new post
│   ├── list           # List posts
│   ├── get            # Get post by slug
│   ├── update         # Update a post
│   └── delete         # Delete a post
└── version            # Show version info
```

**Example Usage**:

```bash
# Seed database with sample data
postctl seed --count 10 --user-id <uuid>

# Posts CRUD (requires API server running)
postctl posts create --title "Hello" --content "World" --user-id <uuid>
postctl posts list --user-id <uuid>
postctl posts get <slug>
postctl posts update <slug> --title "New Title"
postctl posts delete <slug>

# Version
postctl version
```

**Implementation**:

```go
// internal/cli/root.go
package cli

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "postctl",
    Short: "CLI tool for {project-name}",
    Long:  `A command-line interface for managing {project-name} service`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(serverCmd)
    rootCmd.AddCommand(postsCmd)
    rootCmd.AddCommand(versionCmd)
}
```

**Features**:

- Cobra-based CLI framework
- Colored output with lipgloss
- Progress indicators for long operations
- JSON/table output formats
- Configuration file support (~/.postctl/config.yaml)
- Environment variable support
- Verbose/debug modes
- Shell completion (bash, zsh, fish)

**Installation**:

```bash
# Build and install
make cli-install

# Or manually
go build -o /usr/local/bin/postctl cmd/postctl/main.go
```

## 9. Configuration Management

### 9.1 Stage-Based Configuration (YAML + Environment Variables)

**Structure**:

- Config struct in `internal/config/config.go`
- Stage-specific YAML files for non-sensitive configuration
- Environment variables for secrets only
- Type-safe configuration with validation

**Configuration Loading Order**:

1. Load stage-specific YAML file (based on `STAGE` environment variable)
   - `local.yaml` for local development
   - `development.yaml` for development environment
   - `production.yaml` for production (gitignored by default)
2. Load secrets from environment variables (required)

**Key Features**:

- **Stage-Based YAML Files** (`gopkg.in/yaml.v3`):
  - Non-sensitive settings per stage (server port, feature flags, etc.)
  - Easy to read and edit
  - Version controlled (except production.yaml)
  - Stage enum: `local`, `development`, `production`
- **Environment Variables** (secrets only):

  - Secrets (JWT keys, database credentials, API keys)
  - Required for sensitive data
  - Example: `.env` file (gitignored)
  - No override capability - YAML is source of truth for non-secrets

- **Reflection Configuration**:
  - Stage-based: enabled when `STAGE=local`, disabled otherwise
  - No explicit flag - purely stage-based
  - Used for tools like `grpcurl`, `buf`, etc.

**Files Generated**:

- `local.yaml`: Local development configuration (committed)
- `development.yaml`: Development environment configuration (committed)
- `production.yaml`: Production configuration (gitignored by default, but generated)
- `.env.example`: Example environment variables template
- `.env`: Actual secrets (gitignored)

**Example Usage**:

```go
// Load config from YAML with env overrides
cfg, err := config.Load("config.yaml")

// Or load only from environment (useful for containers)
cfg, err := config.LoadFromEnv()
```

## 10. Development Tooling

### 10.1 wgo Hot Reload

**Configuration**:

- `wgo.yaml` file generated
- Watches `internal/`, `cmd/` directories
- Excludes test files
- Build and run commands configured

**Features**:

- Simple YAML configuration
- Fast file watching
- Automatic rebuild and restart
- Minimal overhead

### 10.2 Docker Support

**Files Generated**:

- `Dockerfile`: Multi-stage build
- `docker-compose.yml`: Local development stack
- Includes DynamoDB Local or PostgreSQL for development

## 11. Deployment

### 11.1 Fly.io Integration

**Configuration**:

- `fly.toml` generated with sensible defaults
- Health check endpoint configured
- Environment variables setup
- Build configuration
- HTTP/2 support enabled for gRPC services (`h2_backend = true`)

**Features**:

- Single instance deployment (no staging/prod separation initially)
- Secrets management
- Health checks
- **HTTP/2 Support**: Full support for HTTP/2 and gRPC
  - Edge proxy supports HTTP/2
  - Backend HTTP/2 (h2c) with `h2_backend = true`
  - Perfect for ConnectRPC services

### 11.2 Makefile Deploy Command

**Structure**:

- `Makefile` includes `deploy` target
- Deploys to single Fly.io instance
- Handles build and deployment steps
- Environment variable validation

**Example**:

```makefile
deploy:
	@echo "Deploying to Fly.io..."
	flyctl deploy
```

### 11.3 GitHub Actions CI/CD

**Workflow**:

- Automatic deployment on push to main branch
- Build and test before deployment
- Deploy to Fly.io using GitHub Actions
- Environment secrets management

**Configuration**:

- `.github/workflows/deploy.yml` generated
- Uses Fly.io GitHub Action
- Requires `FLY_API_TOKEN` secret
- Deploys on successful build/test

## 12. Code Generation Logic

### 12.1 Project Generator

**Responsibilities**:

- Create directory structure
- Generate files from templates
- Manage dependencies in `go.mod`
- Validate generated code

### 12.2 Dependency Management

**Approach**:

- Template includes required dependencies
- Generator adds dependencies to `go.mod`
- Version pinning for stability
- Optional dependency resolution

## 13. Future Expansion

### 13.1 Temporal Workflows

**Planned Features**:

- Sample workflow definitions
- Activity implementations
- Worker setup
- Temporal client configuration

### 13.2 Message Queue Integration

**Planned Features**:

- NATS integration
- RabbitMQ integration
- Publisher/subscriber patterns
- Queue configuration

## 14. Testing

### 14.1 Test Strategy

**Structure**:

- Testcontainers for integration tests
- Parallel test execution (`t.Parallel()`)
- Unique directories for file-generating tests
- Comprehensive test coverage

**Key Features**:

- **Parallel Execution**: All tests use `t.Parallel()` for faster execution
- **Testcontainers**:
  - DynamoDB Local for DynamoDB tests
  - PostgreSQL containers for PostgreSQL tests
- **Unique Directories**: Tests that generate files use unique temp directories
- **Build Verification**: Generator tests verify builds, not test execution
- **Migrations**: PostgreSQL tests apply migrations from project root

**Test Organization**:

- Generator tests: Test template embedding, execution, and project generation
- Repository tests: Test database operations with testcontainers
- Handler tests: Test HTTP/gRPC handlers with mocks
- Service tests: Test business logic

## 15. Implementation Phases

### Phase 1: Core CLI & TUI

- Bubbletea TUI implementation
- Basic project scaffolding
- Template system
- Chi REST option

### Phase 2: API Options

- Huma REST option
- gRPC/ConnectRPC option

### Phase 3: Database & Metrics

- DynamoDB integration
- Supabase/PostgreSQL integration
- Metrics instrumentation
- Config management

### Phase 4: Deployment & Tooling

- Fly.io deployment configs
- Makefile with deploy command
- GitHub Actions CI/CD workflow
- wgo hot reload setup
- Docker configuration
- Reflection configuration (local vs production)

### Phase 5: Future Features

- Temporal workflows
- Message queue integration

## 16. Technical Decisions

### 16.1 Why Bubbletea?

- Modern TUI framework with rich component library
- Component-based architecture (Elm-inspired)
- Animated spinners and progress indicators
- Excellent styling with Lipgloss
- Good developer experience
- Active maintenance
- Similar to tools like create-next-app

### 16.2 Why Folder-by-Feature?

- Scalability
- Clear separation of concerns
- Easy to navigate
- Matches current repo structure

### 16.3 Why DynamoDB?

- Serverless-friendly
- No connection pooling needed
- Good for microservices
- Proto serialization support

### 16.4 Why pgx for PostgreSQL?

- Modern, performant driver
- Native struct scanning (no scany dependency needed)
- Type-safe query results with `pgx.RowToStructByName` and `pgx.CollectRows`
- Active development and maintenance
- Built-in connection pooling

### 16.5 Why Atlas Go for Migrations?

- Reproducible migrations
- Up/down migration support
- CLI-based migration creation (not generated)
- Integrated with deployment process

### 16.6 Why Fly.io First?

- Simple deployment
- Good developer experience
- Cost-effective
- Easy to extend to other platforms

### 16.7 Why wgo?

- Simple configuration
- Fast file watching
- Minimal overhead
- Good developer experience

### 16.8 Why Embed Templates?

- Self-contained binary with no external dependencies
- No need to distribute template files separately
- Portable and easy to install (single binary)
- Fast template loading from memory
- Prevents runtime errors from missing template files
- Uses Go's built-in `embed` package (`//go:embed` directive)

### 16.9 Why Stage-Based Configuration?

- Clear separation between environments
- YAML files committed per stage (except production)
- Secrets only in environment variables
- No override complexity - YAML is source of truth

### 16.10 Why PostHog Interface?

- Easy mocking in tests
- Startup validation ensures client is never nil when enabled
- Optional feature - only included when selected
- Clean separation of concerns

## 17. Open Questions

1. Should the CLI support updating existing projects?
2. Should templates be customizable/extensible?
3. Should we support multiple databases in future?
4. Should we include testing templates?
5. Should we support plugin system for custom templates?

## 18. Success Criteria

- CLI generates working microservice boilerplate
- Generated code follows best practices
- All API options work out of the box
- Deployment configs are production-ready
- Documentation is comprehensive
- TUI is intuitive and user-friendly
- Reflection enabled in local stage, disabled in production
- Makefile deploy command works
- GitHub Actions automatically deploys on push to main
- All tests run in parallel
- Testcontainers work for both DynamoDB and PostgreSQL
- Stage-based configuration works correctly
- PostHog integration is optional and properly validated
