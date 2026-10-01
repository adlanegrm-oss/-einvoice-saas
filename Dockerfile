# Étape 1 : construction du binaire Go (sans CGO : pilote SQLite pur Go)
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/server ./cmd/server

# Étape 2 : image d'exécution minimale, sans droits root
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -S app && adduser -S -G app app \
    && mkdir -p /data && chown app:app /data

WORKDIR /app
COPY --from=builder /out/server .
COPY --from=builder /app/web ./web

ENV PORT=8080 \
    DATA_DIR=/data \
    WEB_DIR=/app/web

USER app
VOLUME ["/data"]
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
  CMD wget -qO- "http://127.0.0.1:${PORT}/health" >/dev/null || exit 1

CMD ["./server"]
