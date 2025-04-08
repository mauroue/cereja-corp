# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cereja-corp ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/cereja-corp .
COPY --from=builder /app/config.json ./config.json
COPY --from=builder /app/internal /root/internal

# Create uploads directory
RUN mkdir -p ./uploads/receipts

# Expose port
EXPOSE 8080

# Run the application
CMD ["./cereja-corp"] 