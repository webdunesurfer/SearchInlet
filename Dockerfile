# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /app/mcp-server ./cmd/mcp-server

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/mcp-server .

# Expose port (useful for Phase 2/3 SSE)
EXPOSE 8080

# Run the application
ENTRYPOINT ["./mcp-server"]
