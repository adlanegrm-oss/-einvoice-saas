# Étape 1 : Construction du binaire Go
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache et téléchargement des modules Go
COPY go.mod go.sum* ./
RUN go mod download

# Copie du code source complet
COPY . .

# Compilation statique optimisée sans CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server/main.go

# Étape 2 : Image d'exécution minimale
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copie du binaire et du dossier web
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web

EXPOSE 8080
ENV PORT=8080

CMD ["./server"]
