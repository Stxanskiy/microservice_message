# --------------------------------------------
#       STAGE 1: build
# --------------------------------------------
FROM golang:1.23 as builder

WORKDIR /app
# Копируем .env
COPY .env .env

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o microservice_message ./cmd/main.go

# --------------------------------------------
#       STAGE 2: run
# --------------------------------------------
FROM alpine:3.17

RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow

# Создаем пользователя без root-привилегий (безопаснее)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/microservice_message /app/microservice_message
COPY --from=builder /app/.env /app/.env

# Выставляем права
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8081
ENTRYPOINT ["/app/microservice_message"]
