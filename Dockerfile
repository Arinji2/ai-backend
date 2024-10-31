# Use the official Go image as a build stage
FROM golang:1.23.0-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy and build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Use a minimal image to run the application
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .

# Install necessary packages for profiling
RUN apk add --no-cache curl

# Command to run the application
CMD ["./main"]
