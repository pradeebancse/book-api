# Use the official Golang image to create a build artifact.
FROM golang:1.23-alpine as builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files.
COPY go.mod .
COPY go.sum .

# Download all dependencies.
RUN go mod download

# Copy the source code into the container.
COPY . .

# Build the Go app.
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a minimal alpine image for the final stage.
FROM alpine:latest

# Set the working directory.
WORKDIR /root/

# Copy the pre-built binary file from the previous stage.
COPY --from=builder /app/main .

# Expose port 8080 to the outside world.
EXPOSE 8080

# Command to run the executable.
CMD ["./main"]