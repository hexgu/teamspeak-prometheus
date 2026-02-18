# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o ts3-exporter cmd/ts3-exporter/main.go

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /app/ts3-exporter /ts3-exporter

# Copy example config (optional, but good for reference/mounting)
COPY --from=builder /app/legacy_python/config.yaml.example /config.yaml.example

# Expose the port
EXPOSE 8000

# Command to run
CMD ["/ts3-exporter"]
