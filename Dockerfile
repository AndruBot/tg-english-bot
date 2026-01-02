# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tg-english-bot .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/tg-english-bot .

# Copy questions.json if it exists (optional, can be mounted as volume)
COPY questions.json.example questions.json.example

# Expose port (if needed for health checks)
EXPOSE 8080

# Run the application
CMD ["./tg-english-bot"]

