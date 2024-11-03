# Step 1: Build the Go application with CompileDaemon
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application code
COPY ./cmd/main ./cmd/main 
COPY ./api ./api
# COPY ./internal ./internal
COPY ./config ./config
COPY ./web /app/web            

# Install CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon@latest

# Step 2: Set up the Runtime Container
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary and necessary files from the builder stage
COPY --from=builder /go/bin/CompileDaemon /usr/local/bin/CompileDaemon
# Copy all files from builder
COPY --from=builder /app /app  

# Expose the port
EXPOSE 8003

# Run CompileDaemon to watch for file changes and rebuild/restart the app
ENTRYPOINT ["CompileDaemon", "-build=go build -o main ./cmd/main/main.go", "-command=./main"]
