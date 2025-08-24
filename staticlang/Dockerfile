# StaticLang Compiler Dockerfile
# Multi-stage build for optimized production image

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the compiler
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S staticlang && \
    adduser -u 1001 -S staticlang -G staticlang

# Set working directory
WORKDIR /workspace

# Copy binary from builder stage
COPY --from=builder /app/build/staticlang /usr/local/bin/staticlang

# Set permissions
RUN chmod +x /usr/local/bin/staticlang

# Switch to non-root user
USER staticlang

# Set entrypoint
ENTRYPOINT ["staticlang"]

# Default command shows help
CMD ["-h"]

# Labels
LABEL maintainer="StaticLang Team"
LABEL description="StaticLang Compiler"
LABEL version="0.1.0"