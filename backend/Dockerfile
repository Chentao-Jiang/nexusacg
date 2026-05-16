# syntax=docker/dockerfile:1

# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies first (layer caching)
COPY go.mod go.sum ./
RUN GOPROXY=https://goproxy.cn,direct go mod download

# Copy source and build
COPY . .
RUN GOPROXY=https://goproxy.cn,direct CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install CA certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/server"]
