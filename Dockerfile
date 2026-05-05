FROM alpine:latest

# Add non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the pre-built binary from the host
COPY main .

# Use non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
