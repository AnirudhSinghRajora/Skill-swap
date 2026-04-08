# Use the official Go image as build stage
FROM golang:1.25.9-alpine3.23 AS builder
# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files from the backend/skillswap directory
COPY backend/skillswap/go.mod backend/skillswap/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code from backend/skillswap
COPY backend/skillswap/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Use minimal alpine image for final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh skillswap

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy any static files if needed
COPY --from=builder /app/.env.example .env.example

# Change ownership to skillswap user
RUN chown -R skillswap:skillswap /root/
USER skillswap

# Expose port (Heroku will set $PORT)
EXPOSE 8080

# Command to run the application
CMD ["./main"]
