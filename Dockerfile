# Stage 1: Build.
FROM golang:1.22 AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum and download dependencies.
COPY go.mod go.sum ./

# Download dependencies.
RUN go mod download

# Copy the source code.
COPY . .

# Build the application.
RUN go build -o /app/main .

# Stage 2: Run.
FROM ubuntu:22.04

# Set the working directory inside the container.
WORKDIR /app

# Copy the built application from the builder stage.
COPY --from=builder /app/main .

# Ensure the binary has executable permissions.
RUN chmod +x /app/main

# Expose the port your app runs on.
EXPOSE 80

# Default entry point to pass arguments dynamically.
ENTRYPOINT ["/app/main"]
