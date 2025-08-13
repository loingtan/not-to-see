# Build stage
FROM golang:1.21-alpine AS builder

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cobra-template .

# Final stage
FROM alpine:latest

# Install ca-certificates (needed for HTTPS requests)
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/cobra-template .
COPY --from=builder /app/.cobra-template.yaml .

# Change ownership of the app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./cobra-template server --help || exit 1

# Set environment variables
ENV GIN_MODE=release

# Run the binary
ENTRYPOINT ["./cobra-template"]
CMD ["server"]
