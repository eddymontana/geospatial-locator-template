# Stage 1: Build the Go application binary
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# Stage 2: Final lightweight deployment execution layer
FROM alpine:latest
WORKDIR /app

# 1. Pull the compiled web server binary from the builder layer
COPY --from=builder /app/main .

# 2. Recreate the data directory and stage the spatial dataset
RUN mkdir -p data
COPY --from=builder /app/data/recycling-locations.geojson ./data/

# 3. Expose the standard routing interface port
EXPOSE 8080

# 4. Trigger the server engine
CMD ["./main"]