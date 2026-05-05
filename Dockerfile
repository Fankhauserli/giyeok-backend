# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Copy vendor directory for offline build
COPY vendor ./vendor

# Copy source code
COPY . .

# Build the application as a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main ./cmd/server

# Final stage
FROM alpine:latest

# Add non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
