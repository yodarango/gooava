# Step 1: Build the Go application
FROM golang:1.20-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go binary
RUN go build -o main .

# Step 2: Run the Go application
FROM alpine:3.18

# Set the working directory inside the container
WORKDIR /app

# Copy the Go binary from the builder stage
COPY --from=builder /app/main .

# Copy any other necessary files (such as config files, migrations, etc.)
COPY --from=builder /app/db /app/db

# Expose the port on which the application will run (change if necessary)
EXPOSE 8080

# Define environment variables if needed
# ENV DATABASE_URL=mysql://user:password@tcp(mysql-container:3306)/dbname

# Start the Go application
CMD ["./main"]
