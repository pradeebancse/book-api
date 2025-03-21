# Use Golang base image
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-middleware main.go

# Create a lightweight final image
FROM alpine:latest

WORKDIR /app

# Copy only the built binary from the builder stage
COPY --from=builder /app/auth-middleware .

# Expose the application port
EXPOSE 4000

# Run the service
CMD ["./auth-middleware"]
