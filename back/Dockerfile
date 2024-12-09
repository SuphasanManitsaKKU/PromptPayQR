# Stage 1: Build stage
FROM golang:1.23.3-alpine3.20 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o main .

# Stage 2: Final stage
FROM alpine:latest

# Install necessary libraries for running Go binaries
RUN apk add --no-cache libc6-compat

# Set working directory
WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8000

# Command to run the Go application
CMD ["./main"]
