# ---------------------------------
# 1) Сборка Go-приложения (stage 1)
# ---------------------------------
FROM golang:1.20 AS builder

# Рабочая директория в контейнере
WORKDIR /app

# Копируем go.mod и go.sum, чтобы заранее скачать зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY build .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /messaging_service cmd/main.go

# ---------------------------------
# 2) Финальный контейнер (stage 2)
# ---------------------------------
FROM alpine:3.18

# Создадим некорневую директорию для нашего приложения
WORKDIR /app

# Копируем скомпилированный бинарник из stage1
COPY --from=builder /messaging_service /app/messaging_service

# Порт, на котором запускается сервис
EXPOSE 8081

# Команда запуска
ENTRYPOINT ["/app/messaging_service"]
