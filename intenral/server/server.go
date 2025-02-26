package server

import (
	"context"
	"gitlab.com/nevasik7/lg"
	"log"
	"net/http"
	"sevice_message_1/config"
	"sevice_message_1/intenral/chat/delivery/websocket"
	http2 "sevice_message_1/intenral/chat/handler"
	"sevice_message_1/intenral/chat/repo"
	"sevice_message_1/intenral/chat/uc"
	"sevice_message_1/pkg/cassandra"
	"sevice_message_1/pkg/jwt"
	"sevice_message_1/pkg/kafka"
	"sevice_message_1/pkg/ratelimiter"
	"sevice_message_1/pkg/tracing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupRouter(cfg *config.Config, hub *websocket.Hub) http.Handler {
	//init logger
	lg.Init()

	r := chi.NewRouter()

	// Стандартные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://your-frontend.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	r.Use(ratelimiter.RateLimit(100))
	// OpenTelemetry middleware
	r.Use(tracing.Middleware)

	// Подключение к базе данных PostgreSQL
	pgPool, err := pgxpool.New(context.Background(), cfg.PostgreURL)
	if err != nil {
		log.Fatalf("Unable to connect to Postgres: %v", err)
	}

	// Подключение к Cassandra
	cassandraSession, err := cassandra.NewSession(cfg.CassandraURL)
	if err != nil {
		log.Fatalf("Unable to connect to Cassandra: %v", err)
	}

	// Инициализация Kafka-продюсера
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatalf("Unable to connect to Kafka: %v", err)
	}

	// Инициализация репозитория и бизнес-логики (use case) для чатов
	chatRepo := repo.NewChatRepo(pgPool, cassandraSession)
	if chatRepo == nil {
		lg.Errorf("ChatRepo is nil %v", err)
	}
	chatUC := uc.NewChatUC(chatRepo, kafkaProducer)
	if chatUC == nil {
		lg.Errorf("ChatUC is nil %v", err)
	}

	// Инициализация JWT менеджера (только для проверки токенов)
	jwtManager := jwt.NewJWTManager(cfg.JwtSecret)
	authMiddleware := jwt.AuthMiddleware(jwtManager)

	// Регистрируем приватные эндпоинты для чатов, используя созданный маршрутизатор,
	// который возвращается функцией NewRouter (она возвращает объект, реализующий http.Handler)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Mount("/api/chats", http2.NewRouter(chatUC))
		r.Handle("/ws/chats/{chatID}", websocket.NewWebSocketHandler(chatUC, hub))
	})
	return r
}

func StartServer(cfg *config.Config) {

	//Инициализируем hub
	hub := websocket.NewHub()
	go hub.Run()

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      SetupRouter(cfg, hub),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	lg.Infof("Messaging service starting on port %s", cfg.ServerPort+"\nLink:http://localhost:8081")
	if err := srv.ListenAndServe(); err != nil {
		lg.Fatalf("Server error: %v", err)

	}
}
