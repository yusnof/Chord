FROM golang:1.25-alpine AS build

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./


# Copy source code
COPY . .

# Build the binary
RUN go build -o chord .

FROM alpine:latest

WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/chord .

# Default command (will be overridden by docker-compose)
CMD ["./chord"]
