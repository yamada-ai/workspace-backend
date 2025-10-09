# ---------- build stage ----------
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /work-tracker ./cmd/work-tracker

# ---------- runtime stage ----------
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /work-tracker .

# Copy migrations (for reference)
COPY migrations ./migrations

EXPOSE 8000

CMD ["./work-tracker"]
