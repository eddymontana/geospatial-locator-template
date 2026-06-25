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

# 2. Recreate and stage the static frontend folder layout
RUN mkdir -p static
COPY --from=builder /app/static/ ./static/

# 3. Recreate and stage the geospatial data asset layout
RUN mkdir -p data
COPY --from=builder /app/data/recycling-locations.geojson ./data/

# 4. Expose the standard routing interface port
EXPOSE 8080

# 5. Trigger the server engine
CMD ["./main"]