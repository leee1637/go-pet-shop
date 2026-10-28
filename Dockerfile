# Этап сборки
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o go-pet-shop ./cmd/app
RUN go build -o migrator ./cmd/migrator

# Финальный минимальный образ
FROM alpine:3.20.3

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/go-pet-shop .
COPY --from=builder /app/migrator .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./go-pet-shop"]
