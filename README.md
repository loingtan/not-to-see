# Cobra Template

A well-structured Go project template using Cobra CLI framework with clean architecture, mock endpoints, and comprehensive documentation.

## 🏗️ Architecture

This project follows Domain-Driven Design (DDD) principles and clean architecture patterns:

```
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── api/              # HTTP layer (handlers, middleware, router)
│   ├── config/           # Configuration management
│   ├── domain/           # Business logic and entities
│   ├── infrastructure/   # External concerns (database, cache, etc.)
│   └── service/          # Application services
├── pkg/                  # Public packages
├── docs/                 # Documentation
├── examples/             # Usage examples
└── scripts/              # Build and deployment scripts
```

## 🚀 Features

- **Clean Architecture**: Separation of concerns with clear boundaries
- **Domain-Driven Design**: Business logic at the core
- **RESTful API**: Well-designed HTTP endpoints with proper status codes
- **Middleware**: Request logging, CORS, recovery
- **Configuration**: Flexible configuration with Viper
- **Logging**: Structured logging with Logrus
- **Validation**: Request validation with go-playground/validator
- **Mock Data**: In-memory repository for testing and development
- **Health Checks**: Health, readiness, and liveness endpoints
- **Graceful Shutdown**: Proper server shutdown handling

## 🛠️ Quick Start

### Prerequisites

- Go 1.21+
- Docker + Docker Compose (optional, for Postgres/Redis)
- Make (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd cobra-template
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o bin/cobra-template
```

### Running the Application

#### Start the HTTP server:
```bash
go run main.go registration
# or
./bin/cobra-template registration
```

#### With custom port:
```bash
go run main.go registration --port 3000
```

#### With verbose logging:
```bash
go run main.go registration --verbose
```

### Using Docker Compose (recommended for local DB/Cache)

Start Postgres and Redis locally via Docker:

```bash
make up
# or
docker compose up -d postgres redis
```

Then run the app (from your host):

```bash
make run
# or
./bin/cobra-template registration -v
```

Stop services when done:

```bash
make down
# or
docker compose down
```



## 📚 API Documentation

The server provides a REST API with the following endpoints:

### Health Checks
- `GET /health` - Application health status
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe

### Users API
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List users (with pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `GET /api/v1/users/email/:email` - Get user by email
- `GET /api/v1/users/username/:username` - Get user by username

### API Documentation
- `GET /docs` - Interactive API documentation

## 🧪 Examples

### Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "first_name": "New",
    "last_name": "User"
  }'
```

### Get Users (with pagination)
```bash
curl "http://localhost:8080/api/v1/users?limit=5&offset=0"
```

### Update a User
```bash
curl -X PUT http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Updated Name"
  }'
```

## ⚙️ Configuration

The application can be configured using:

1. **Configuration file**: `.cobra-template.yaml` (YAML format)
2. **Environment variables**: Prefixed with app name
3. **Command-line flags**: Available flags depend on the command

### Configuration Structure

See `.cobra-template.yaml` for the complete configuration structure.

### Environment Variables

All configuration values can be overridden using environment variables:

```bash
export APP_NAME="my-app"
export SERVER_PORT="3000"
export LOG_LEVEL="debug"
```

## 🏗️ Development

### Project Structure

- **`cmd/`**: Contains the CLI commands and main application entry points
- **`internal/api/`**: HTTP layer with handlers, middleware, and routing
- **`internal/domain/`**: Core business entities and interfaces
- **`internal/service/`**: Business logic implementation
- **`internal/infrastructure/`**: External dependencies (databases, caches, etc.)
- **`pkg/`**: Reusable packages that could be imported by other projects

### Adding New Features

1. **Define Domain**: Add entities and interfaces in `internal/domain/`
2. **Implement Service**: Add business logic in `internal/service/`
3. **Create Repository**: Add data access in `internal/infrastructure/`
4. **Add Handlers**: Create HTTP handlers in `internal/api/handlers/`
5. **Update Router**: Register routes in `internal/api/router/`

### Testing

```bash
go test ./...
```

### Building

```bash
# Development build
go build -o bin/cobra-template

# Production build with optimizations
go build -ldflags="-w -s" -o bin/cobra-template
```

## 🐳 Docker Support

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-w -s" -o bin/cobra-template

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/cobra-template .
COPY --from=builder /app/.cobra-template.yaml .
EXPOSE 8080
CMD ["./cobra-template", "server"]
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🔗 Resources

- [Cobra CLI](https://github.com/spf13/cobra)
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Viper Configuration](https://github.com/spf13/viper)
- [Logrus Logging](https://github.com/sirupsen/logrus)
- [Go Validator](https://github.com/go-playground/validator)
