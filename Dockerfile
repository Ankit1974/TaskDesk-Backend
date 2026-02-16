# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first (for better Docker layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o taskdesk-api ./cmd/api

# ---- Run Stage ----
FROM alpine:3.21

WORKDIR /app

# Install ca-certificates (needed for HTTPS calls to Supabase)
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/taskdesk-api .

# Copy migrations (if you run them at startup or manually)
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./taskdesk-api"]
