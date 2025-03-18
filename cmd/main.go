package main

import (
	"github.com/joho/godotenv"
	"gitlab.com/nevasik7/lg"
	"sevice_message_1/config"
	"sevice_message_1/intenral/server"
)

func main() {
	// Init logger
	lg.Init()
	//env
	if err := godotenv.Load(".env"); err != nil {
		lg.Errorf("No .env file found, using system environment variables %v", err)

	}

	cfg, err := config.LoadCOnfig()
	if err != nil {
		lg.Errorf("Error loading config: %v", err)
	}

	// Инициализируем распределённое трассирование (Jaeger)
	//TODO Потом влючить jaeger обратно
	/*tp, err := tracing.InitTracer(cfg.JaegerURL, "messaging_service")
	if err != nil {
		lg.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			lg.Fatalf("Error shutting down tracer: %v", err)
		}
	}()*/

	//TODO он инициализирован уже в StartServer
	////Инициализации вебсокет хабл
	//var hub *websocket.Hub

	server.StartServer(cfg)

	//TODO Если что можно вернуть обратно при некоректной работы сервиса
	//srv := &http.Server{
	//	Addr:         ":" + cfg.ServerPort,
	//	Handler:      server.SetupRouter(cfg, hub),
	//	ReadTimeout:  15 * time.Second,
	//	WriteTimeout: 15 * time.Second,
	//}
	//
	//lg.Infof("Messaging service starting on port %s", cfg.ServerPort)
	//if err := srv.ListenAndServe(); err != nil {
	//	lg.Fatalf("Server error: %v", err)
	//}

}
