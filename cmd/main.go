package main

import (
	"github.com/go-chi/chi/v5"
	"gitlab.com/nevasik7/lg"
	"serice_message/config"
	"serice_message/internal/chat/handler"
	"serice_message/internal/chat/route"
	"serice_message/internal/chat/uc"
	"serice_message/pkg/jwt"
	"serice_message/pkg/kafka"

	"net/http"
	"time"
)

func main() {
	//загрузка конфигурации
	lg.Init()
	cfg, err := config.LoadConfig()
	if err != nil {
		lg.Fatalf("Failed to load config:%v", err)
	}

	//database connected
	db, err := config.NewDatabaseConnected(cfg.DatabaseURL)
	if err != nil {
		lg.Fatalf("Failed to conect to database:%v", err)
	}
	defer db.Close()

	// TODO: Добавить запуск WebSocket-сервера, Kafka и и маршрутов HTTP

	//JWT MAnager
	jwtManger := jwt.NewJWTManager("salt_secret", 15*time.Minute)

	//Kafka Producer
	kafkaProducer := kafka.NewKafkaProducer(cfg.KafkaBrokers, cfg.MessageTopic)
	defer kafkaProducer.Close()

	messageUC := uc.NewMessageUC(kafkaProducer)

	//WebSocket Handler
	wsHandler := handler.NewWebSocketHandler(jwtManger, messageUC)

	//route
	r := chi.NewRouter()
	route.RegisterWebSocketRoutes(r, wsHandler)

	//запуск сервера
	lg.Infof("Startin server on linl http://localhost:8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		lg.Fatalf("Failed  to start server:%v, err")
	}
}
