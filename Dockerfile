# syntax=docker/dockerfile:1

# Stage 1: Build Next.js static export
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go single binary
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

# Copy backend source and embedded seed SQL
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
COPY seeds/ ./seeds/
COPY migrations/ ./migrations/
COPY frontend/dist.go ./frontend/dist.go

# Copy static frontend output from Stage 1 into frontend/out
COPY --from=frontend-builder /app/frontend/out ./frontend/out

# Static compilation (pure Go SQLite modernc.org/sqlite, no CGO required)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 3: Minimal secure production runtime
FROM alpine:3.21 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 10001 -S penlight \
    && adduser -u 10001 -S penlight -G penlight \
    && mkdir -p /data && chown -R penlight:penlight /data /app

# Copy single binary from backend builder
COPY --from=backend-builder --chown=penlight:penlight /app/server /app/server

USER penlight:penlight

ENV DATA_DIR=/data
EXPOSE 8080

ENTRYPOINT ["/app/server"]
