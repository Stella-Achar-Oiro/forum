# Use the official Go image as the base image
FROM golang:1.24.1-alpine

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Create necessary directories
RUN mkdir -p /app/static /app/database /app/uploads

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o main .

# Ensure proper permissions
RUN chmod -R 755 /app/static /app/database /app/uploads

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./main"] 