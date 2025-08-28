#
FROM golang:1.24.4-alpine AS builder


RUN apk add --no-cache git ca-certificates tzdata


WORKDIR /app


COPY go.mod go.sum ./


RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest


RUN apk --no-cache add ca-certificates tzdata


RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup


WORKDIR /app


COPY --from=builder /app/main .


COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations


RUN mkdir -p logs && chown -R appuser:appgroup /app


USER appuser


EXPOSE 8080


HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1


ENV GIN_MODE=release
ENV CONFIG_PATH=/app/configs


CMD ["./main", "registration", "--port", "8080"]
