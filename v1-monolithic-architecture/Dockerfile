# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o api-server ./cmd/api-server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/api-server .

# Copy necessary files
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example ./.env.example

# Create directories for uploads and storage
RUN mkdir -p uploads storage

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-server"]
