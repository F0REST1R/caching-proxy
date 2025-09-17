# 1. Этап сборки (builder)
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 2. Копируем только файлы зависимостей (для кэширования)
COPY go.mod go.sum ./
RUN go mod download

# 3. Копируем весь код и собираем приложение
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /caching-proxy ./cmd 

# 4. Этап запуска (минимальный образ)
FROM alpine:3.21.3

WORKDIR /

# 5. Копируем бинарник и сертификаты
COPY --from=builder /caching-proxy /caching-proxy
RUN apk add --no-cache ca-certificates

# 6. Указываем точку входа
EXPOSE 3000
ENTRYPOINT ["/caching-proxy"]

# docker-compose запускается по команде docker-compose --env-file docker.env up
