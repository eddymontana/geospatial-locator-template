# Stage 1: Build the Go binary
FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# Stage 2: Final lightweight execution layer
FROM alpine:latest
WORKDIR /app

# 1. Copy the compiled binary
COPY --from=builder /app/main .

# 2. CRITICAL: Copy the spatial data file into the execution directory!
COPY --from=builder /app/recycling-locations.geojson .

EXPOSE 8080
CMD ["./main"]