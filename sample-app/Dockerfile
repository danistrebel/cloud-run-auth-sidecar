# Stage 1: Build the Go application
FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o main .

# Stage 2: Create a minimal image from scratch
FROM scratch

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main /main

# Copy the truststore
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Expose the port the application listens on
EXPOSE 8080

# Command to run the application
CMD ["/main"]