# BINARY selects which cmd/* to build. Defaults to api-server.
# Override per-service in docker-compose: build.args.BINARY=payment-worker
ARG BINARY=api-server

# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25.3-alpine AS builder
ARG BINARY

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/bin/${BINARY} ./cmd/${BINARY}

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:latest
ARG BINARY
ENV BINARY=${BINARY}

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/bin/${BINARY} ./service
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example ./.env.example

RUN mkdir -p uploads storage

EXPOSE 8080

CMD ["./service"]
