FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

# Dependencies first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/cobra-template

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /out/cobra-template ./cobra-template
COPY configs/.cobra-template.yaml ./configs/.cobra-template.yaml

EXPOSE 8080
CMD ["./cobra-template", "registration", "--config", "./configs/.cobra-template.yaml"]

