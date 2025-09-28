# DynamoDB Note-Taking App

A RESTful API for note-taking with CRUD operations, built with Go, AWS SDK v2, and DynamoDB. Features both public and private notes with JWT authentication.

## Features

- 📝 **CRUD Operations**: Create, read, update, and delete notes
- 🔐 **JWT Authentication**: Secure private notes with JWT tokens
- 🏠 **Local Development**: Run with DynamoDB Local for development
- 🧪 **Comprehensive Testing**: Unit tests with DynamoDB mocks
- 🐳 **Docker Support**: Containerized application with Docker Compose
- ⚙️ **Configuration**: Environment-based configuration with carlos/env

## API Endpoints

### Public Endpoints (No Authentication Required)

- `GET /health` - Health check
- `GET /notes` - List all public notes
- `POST /notes` - Create a new note (public by default)
- `GET /notes/{id}` - Get a specific note
- `PUT /notes/{id}` - Update a note
- `DELETE /notes/{id}` - Delete a note

### Private Endpoints (JWT Authentication Required)

- `GET /private/notes` - List user's private notes
- `POST /private/notes` - Create a new private note
- `GET /private/notes/{id}` - Get a specific private note
- `PUT /private/notes/{id}` - Update a private note
- `DELETE /private/notes/{id}` - Delete a private note

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and start the services:**
   ```bash
   git clone <repository-url>
   cd create-go-service
   docker-compose up -d
   ```

2. **Test the API:**
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Create a public note
   curl -X POST http://localhost:8080/notes \
        -H "Content-Type: application/json" \
        -d '{"title": "My First Note", "content": "This is a public note"}'
   
   # List notes
   curl http://localhost:8080/notes
   ```

### Local Development

1. **Start DynamoDB Local:**
   ```bash
   docker run -p 8000:8000 amazon/dynamodb-local:2.0.0 -jar DynamoDBLocal.jar -sharedDb -inMemory
   ```

2. **Set environment variables:**
   ```bash
   export DYNAMODB_ENDPOINT=http://localhost:8000
   export TABLE_NAME=notes
   export JWT_SECRET=your-secret-key
   export AWS_REGION=us-east-1
   ```

3. **Run the application:**
   ```bash
   go mod tidy
   go run main.go
   ```

## Configuration

The application uses environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `JWT_SECRET` | `your-secret-key` | JWT signing secret |
| `AWS_REGION` | `us-east-1` | AWS region |
| `TABLE_NAME` | `notes` | DynamoDB table name |
| `DYNAMODB_ENDPOINT` | `""` | DynamoDB endpoint (for local development) |

## JWT Authentication

### Generating a Test Token

```go
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

func generateTestToken(secret string) string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": "test-user-123",
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })
    
    tokenString, _ := token.SignedString([]byte(secret))
    return tokenString
}
```

### Using the Token

```bash
# Get a token (replace with your generated token)
TOKEN="your-jwt-token-here"

# Create a private note
curl -X POST http://localhost:8080/private/notes \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"title": "Private Note", "content": "This is private", "is_private": true}'

# List private notes
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/private/notes
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
```

### Test Categories

- **Repository Tests**: Test DynamoDB operations with mocks
- **Handler Tests**: Test HTTP endpoints with mock repositories
- **Middleware Tests**: Test JWT authentication middleware

## Project Structure

```
├── main.go                          # Application entry point
├── docker-compose.yml               # Local development setup
├── Dockerfile                       # Container configuration
├── go.mod                           # Go module dependencies
├── internal/
│   ├── config/                      # Configuration management
│   │   └── config.go
│   ├── models/                      # Data models
│   │   └── note.go
│   ├── repository/                  # DynamoDB repository
│   │   ├── dynamodb.go
│   │   └── dynamodb_test.go
│   ├── handlers/                    # HTTP handlers
│   │   ├── notes.go
│   │   └── notes_test.go
│   └── middleware/                  # HTTP middleware
│       ├── auth.go
│       └── auth_test.go
```

## DynamoDB Table Schema

The application expects a DynamoDB table with the following structure:

- **Primary Key**: `id` (String)
- **Attributes**:
  - `title` (String)
  - `content` (String)
  - `user_id` (String)
  - `is_private` (Boolean)
  - `created_at` (String, ISO 8601)
  - `updated_at` (String, ISO 8601)

## Development

### Adding New Features

1. **Models**: Add new fields to `internal/models/note.go`
2. **Repository**: Implement new methods in `internal/repository/dynamodb.go`
3. **Handlers**: Add new endpoints in `internal/handlers/notes.go`
4. **Tests**: Write comprehensive tests for new functionality

### Code Style

- Use `gofmt` for formatting
- Follow Go naming conventions
- Write tests for all public functions
- Use meaningful variable and function names

## Production Deployment

### AWS Deployment

1. **Create DynamoDB Table** (using Terraform or AWS Console)
2. **Set Environment Variables**:
   ```bash
   export DYNAMODB_ENDPOINT=""  # Use AWS DynamoDB
   export TABLE_NAME=your-table-name
   export JWT_SECRET=your-production-secret
   export AWS_REGION=your-region
   ```
3. **Deploy**: Use your preferred deployment method (ECS, Lambda, EC2, etc.)

### Security Considerations

- Use strong JWT secrets in production
- Enable HTTPS
- Implement rate limiting
- Use AWS IAM roles for DynamoDB access
- Monitor and log all operations

## Troubleshooting

### Common Issues

1. **DynamoDB Connection Issues**:
   - Ensure DynamoDB Local is running on port 8000
   - Check `DYNAMODB_ENDPOINT` environment variable

2. **Authentication Issues**:
   - Verify JWT secret matches between token generation and server
   - Check token expiration time
   - Ensure proper Bearer token format

3. **Table Not Found**:
   - Create the DynamoDB table with the correct schema
   - Verify table name in configuration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

MIT License - see LICENSE file for details.
