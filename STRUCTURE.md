# Project Structure

This document provides an overview of the Cobra Template project structure and the purpose of each component.

## Directory Layout

```
cobra-template/
├── cmd/                                    # Command-line interface
│   ├── root.go                            # Root command and CLI setup
│   └── server.go                          # Server command implementation
├── internal/                               # Private application code
│   ├── api/                               # HTTP API layer
│   │   ├── handlers/                      # HTTP request handlers
│   │   │   ├── health_handler.go          # Health check endpoints
│   │   │   └── user_handler.go            # User management endpoints
│   │   ├── middleware/                    # HTTP middleware
│   │   │   ├── cors.go                    # Cross-Origin Resource Sharing
│   │   │   └── logger.go                  # Request logging middleware
│   │   └── router/                        # Route configuration
│   │       └── router.go                  # Main router setup
│   ├── config/                            # Configuration management
│   │   └── config.go                      # Configuration structure and loading
│   ├── domain/                            # Business domain layer
│   │   └── user.go                        # User entity and interfaces
│   ├── infrastructure/                    # External dependencies
│   │   └── repository/                    # Data access layer
│   │       └── mock_user_repository.go    # In-memory mock repository
│   └── service/                           # Business logic layer
│       ├── user_service.go                # User business logic
│       └── user_service_test.go           # User service tests
├── pkg/                                   # Public packages
│   ├── logger/                           # Logging utilities
│   │   └── logger.go                     # Structured logging implementation
│   └── validator/                        # Validation utilities
│       └── validator.go                  # Request validation helpers
├── docs/                                 # Documentation
│   └── API.md                           # API documentation
├── examples/                             # Usage examples
│   └── api_examples.sh                  # Shell script with API examples
├── bin/                                 # Built binaries (generated)
├── .cobra-template.yaml                 # Default configuration file
├── .gitignore                          # Git ignore rules
├── Dockerfile                          # Docker container definition
├── Makefile                            # Build and development tasks
├── README.md                           # Main project documentation
├── STRUCTURE.md                        # This file
├── go.mod                              # Go module definition
├── go.sum                              # Go module checksums
└── main.go                             # Application entry point
```

## Architecture Overview

### Clean Architecture Layers

1. **Domain Layer** (`internal/domain/`)
   - Core business entities and interfaces
   - Independent of external concerns
   - Contains business rules and logic

2. **Service Layer** (`internal/service/`)
   - Application business logic
   - Implements domain interfaces
   - Orchestrates between different components

3. **Infrastructure Layer** (`internal/infrastructure/`)
   - External dependencies (databases, caches, etc.)
   - Implements domain interfaces for data access
   - Contains technology-specific implementations

4. **API Layer** (`internal/api/`)
   - HTTP handlers and routing
   - Request/response transformation
   - Middleware for cross-cutting concerns

5. **Configuration** (`internal/config/`)
   - Application configuration management
   - Environment-specific settings
   - Centralized configuration access

6. **Packages** (`pkg/`)
   - Reusable utilities
   - Can be imported by external projects
   - Common functionality (logging, validation)

### Design Patterns Used

1. **Dependency Injection**
   - Services depend on interfaces, not concrete types
   - Easy to test and swap implementations

2. **Repository Pattern**
   - Abstracts data access logic
   - Provides clean interface for data operations

3. **Service Pattern**
   - Encapsulates business logic
   - Provides transaction boundaries

4. **Middleware Pattern**
   - Cross-cutting concerns (logging, CORS, etc.)
   - Composable request processing pipeline

5. **Command Pattern**
   - CLI commands are self-contained
   - Easy to add new commands and features

## Key Components

### Commands (`cmd/`)

- **root.go**: Defines the base CLI structure, global flags, and configuration loading
- **server.go**: HTTP server command with graceful shutdown and signal handling

### Handlers (`internal/api/handlers/`)

- **health_handler.go**: Provides health, readiness, and liveness probes
- **user_handler.go**: Complete CRUD operations for user management

### Middleware (`internal/api/middleware/`)

- **cors.go**: Handles Cross-Origin Resource Sharing for web applications
- **logger.go**: Structured request logging with performance metrics

### Domain (`internal/domain/`)

- **user.go**: User entity, value objects, and repository/service interfaces

### Services (`internal/service/`)

- **user_service.go**: Business logic for user operations with validation and error handling

### Infrastructure (`internal/infrastructure/`)

- **mock_user_repository.go**: In-memory implementation for development and testing

### Utilities (`pkg/`)

- **logger.go**: Structured logging with configurable output formats
- **validator.go**: Request validation with detailed error messages

## Configuration

The application supports multiple configuration sources:

1. **Configuration Files**: YAML format (`.cobra-template.yaml`)
2. **Environment Variables**: Automatic mapping with prefixes
3. **Command-line Flags**: Override specific configuration values

Configuration precedence (highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

## Testing Strategy

- **Unit Tests**: Business logic testing with mocked dependencies
- **Integration Tests**: Full request/response cycle testing
- **Mock Implementations**: In-memory repositories for fast testing

## Development Workflow

1. **Domain First**: Define business entities and interfaces
2. **Service Implementation**: Implement business logic
3. **Infrastructure**: Add data access implementations
4. **API Layer**: Create HTTP handlers and routes
5. **Testing**: Add comprehensive test coverage

## Deployment Options

1. **Binary Deployment**: Single binary with embedded configuration
2. **Docker Container**: Multi-stage build with Alpine Linux
3. **Kubernetes**: Health checks and graceful shutdown support

## Extension Points

1. **New Entities**: Add to `internal/domain/`
2. **New Services**: Add to `internal/service/`
3. **New Handlers**: Add to `internal/api/handlers/`
4. **New Commands**: Add to `cmd/`
5. **New Middleware**: Add to `internal/api/middleware/`

This structure provides a solid foundation for building scalable, maintainable Go applications while following industry best practices and clean architecture principles.
