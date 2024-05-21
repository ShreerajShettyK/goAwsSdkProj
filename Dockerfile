# Use the official Golang image as the base image
FROM golang:1.18-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal image as the base image for the final container
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary file from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/db /root/db

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
