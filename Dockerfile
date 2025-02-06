# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o forum ./cmd/forum

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/forum .

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/internal/database/schema.sql ./internal/database/schema.sql

# Create data directory
RUN mkdir -p /app/data && \
    chmod 777 /app/data

# Expose port
EXPOSE 8080

# Run the application
CMD ["./forum"] 