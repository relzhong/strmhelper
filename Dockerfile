# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO_ENABLED=1 for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o strmhelper ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite

# Create config directory for database persistence
RUN mkdir -p /app/config

# Copy binary from builder
COPY --from=builder /app/strmhelper .

# Copy UI assets
COPY ui ./ui

# Set environment variables
ENV TZ=Asia/Shanghai
ENV STRM_CONFIG_DIR=/app/config

# Expose port
EXPOSE 8080

# Command to run
CMD ["./strmhelper"]
