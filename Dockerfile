# Use the official Golang image as the base image
FROM golang:1.18-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY . .

# Copy the aws_credentials file to the container
# COPY aws_credentials /root/.aws/credentials

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container

# Build the Go app
RUN go build -o main .
RUN go build -o createMongodb ./db/createMongodb.go

# Use a minimal image as the base image for the final container
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary files from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/createMongodb ./db/createMongodb

# Copy the .env file
COPY .env /root/.env

# Copy the shell script
COPY start.sh /root/start.sh

# Debugging step to list the contents of the /root directory
RUN ls -la /root/

# Ensure the correct line endings
RUN sed -i 's/\r$//' /root/start.sh

# Make the shell script executable
RUN chmod +x /root/start.sh

# Expose necessary ports
EXPOSE 8000

# Command to run the shell script
CMD ["/root/start.sh"]
