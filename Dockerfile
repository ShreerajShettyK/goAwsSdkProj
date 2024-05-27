# Use the official Golang image as the base image
FROM golang:1.18-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
# COPY go.mod go.sum ./
COPY . .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container

# Build the Go app
RUN go build -a -installsuffix cgo -o main .

# Build the createMongodb.go web service
RUN go build -a -installsuffix cgo -o createMongodb ./db/createMongodb.go

# Use a minimal image as the base image for the final container
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/createMongodb ./db/createMongodb

# Copy the .env file
COPY ./db/.env ./db/.env

# Expose necessary ports
EXPOSE 8000
# EXPOSE 8081  

# Command to run both services
CMD ./main & ./db/createMongodb
