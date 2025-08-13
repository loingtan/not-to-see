# API Documentation

This document provides comprehensive information about the REST API endpoints available in the Cobra Template application.

## Base URL

```
http://localhost:8080
```

## Authentication

Currently, the API does not require authentication. In a production environment, you would typically implement JWT tokens, API keys, or other authentication mechanisms.

## Common Response Format

All API endpoints follow a consistent response format:

```json
{
  "success": true|false,
  "message": "Optional message",
  "data": {}, // Response data (when applicable)
  "errors": [] // Validation errors (when applicable)
}
```

## Error Handling

The API uses standard HTTP status codes:

- `200` - OK
- `201` - Created
- `400` - Bad Request
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

## Health Check Endpoints

### GET /health

Returns the overall health status of the application and its dependencies.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "cache": "healthy"
  }
}
```

### GET /ready

Readiness probe - indicates if the application is ready to serve traffic.

**Response:**
```json
{
  "ready": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /live

Liveness probe - indicates if the application is running.

**Response:**
```json
{
  "alive": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## User Management API

### POST /api/v1/users

Creates a new user.

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john.doe@example.com",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Validation Rules:**
- `username`: required, 3-50 characters
- `email`: required, valid email format
- `first_name`: required, 1-50 characters
- `last_name`: required, 1-50 characters

**Response (201):**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Response (400):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "tag": "email",
      "message": "email must be a valid email address"
    }
  ]
}
```

### GET /api/v1/users

Lists users with pagination support.

**Query Parameters:**
- `limit` (optional): Number of users to return (default: 10, max: 100)
- `offset` (optional): Number of users to skip (default: 0)

**Example Request:**
```
GET /api/v1/users?limit=5&offset=10
```

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "johndoe",
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "active": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### GET /api/v1/users/:id

Retrieves a specific user by ID.

**Path Parameters:**
- `id`: UUID of the user

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "message": "user not found"
}
```

### PUT /api/v1/users/:id

Updates an existing user. All fields are optional.

**Path Parameters:**
- `id`: UUID of the user

**Request Body:**
```json
{
  "username": "newusername",
  "email": "newemail@example.com",
  "first_name": "NewFirstName",
  "last_name": "NewLastName",
  "active": false
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "newusername",
    "email": "newemail@example.com",
    "first_name": "NewFirstName",
    "last_name": "NewLastName",
    "active": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

### DELETE /api/v1/users/:id

Deletes a user.

**Path Parameters:**
- `id`: UUID of the user

**Response (200):**
```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

### GET /api/v1/users/email/:email

Retrieves a user by email address.

**Path Parameters:**
- `email`: Email address of the user

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### GET /api/v1/users/username/:username

Retrieves a user by username.

**Path Parameters:**
- `username`: Username of the user

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

## cURL Examples

Here are some practical examples using cURL:

### Create a User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice123",
    "email": "alice@example.com",
    "first_name": "Alice",
    "last_name": "Johnson"
  }'
```

### List Users
```bash
curl "http://localhost:8080/api/v1/users?limit=5&offset=0"
```

### Get User by ID
```bash
curl "http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000"
```

### Update User
```bash
curl -X PUT http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000 \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alicia",
    "active": false
  }'
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/api/v1/users/123e4567-e89b-12d3-a456-426614174000
```

### Get User by Email
```bash
curl "http://localhost:8080/api/v1/users/email/alice@example.com"
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Rate Limiting

Currently, no rate limiting is implemented. In a production environment, you would typically implement rate limiting based on IP address or API key.

## Pagination

For endpoints that return lists (like `/api/v1/users`), pagination is supported through query parameters:

- `limit`: Maximum number of items to return (default: 10)
- `offset`: Number of items to skip (default: 0)

Example:
```
GET /api/v1/users?limit=20&offset=40
```

This would return items 41-60 from the result set.

## Response Times

All endpoints are designed to respond within:
- Health checks: < 50ms
- User operations: < 200ms
- List operations: < 500ms

## WebSocket Support

WebSocket endpoints are not currently implemented but can be added for real-time features.

## File Upload

File upload endpoints are not currently implemented but can be added using multipart/form-data.
