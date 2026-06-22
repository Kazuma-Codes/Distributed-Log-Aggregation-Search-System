# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately from application code.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o /query-api .

# =============================================================================
# Run stage — minimal image with no build toolchain
# =============================================================================
FROM alpine:3.20

# CA certificates are needed for TLS connections to external services.
RUN apk --no-cache add ca-certificates

COPY --from=builder /query-api /query-api

EXPOSE 8080

ENTRYPOINT ["/query-api"]
