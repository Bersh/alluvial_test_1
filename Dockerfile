FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o eth-balance-proxy ./cmd/server

# Use a small alpine image for the final container
FROM alpine:3.18

WORKDIR /app

# Install curl for health checks
RUN apk --no-cache add curl

# Copy the binary from builder
COPY --from=builder /app/eth-balance-proxy .

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./eth-balance-proxy"]