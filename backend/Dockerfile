# ==========================================
# Multi-stage Dockerfile for Eka-Hub Go Backend
# ==========================================

# Stage 1: Build binary
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY . .

# Build statically-linked optimized binary
ARG APP_VERSION=""
ARG ENVIRONMENT="production"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /build/bin/server \
    cmd/server/main.go

# Stage 2: Minimal runtime image
FROM alpine:3.20 AS runner

WORKDIR /app

# Add runtime certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy compiled binary from builder
COPY --from=builder /build/bin/server /app/server
# Copy migrations for startup auto-migration / tooling
COPY --from=builder /build/migrations /app/migrations

USER appuser

EXPOSE 3000

ENV ENVIRONMENT=production
ENV SERVER_PORT=3000

ENTRYPOINT ["/app/server"]
