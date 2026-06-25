# Step 1: Build Stage (Compiles the Go application)
FROM golang:alpine AS builder

# Set the working directory inside the build container
WORKDIR /app

# Copy dependency files first for efficient caching layers
COPY go.mod ./
RUN [ -f go.sum ] && COPY go.sum ./ || echo "No go.sum yet"
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build a purely static, highly optimized binary for Linux targets
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -o geospatial-app .

# Step 2: Final Runtime Stage (Minimal and Secure)
FROM alpine:latest  

# Install CA certificates so your Go backend can make secure HTTPS requests to Google Maps API
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the static frontend assets directly into the application runtime root
COPY --from=builder /app/static ./static

# Copy the compiled application binary from the builder stage
COPY --from=builder /app/geospatial-app .

# Expose the port Cloud Run routes traffic to dynamically
EXPOSE 8080

# Run the Go binary from the root directory context
CMD ["./geospatial-app"]