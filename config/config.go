package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gitlab.com/nevasik7/lg"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
	MessageTopic string
	KafkaUI      string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		lg.Panicf("Error loading .env file: %v", err)
	}

	// Считываем переменные окружения
	databaseURL := os.Getenv("DATABASE_URL")
	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS") // Например: "localhost:9092"
	messageTopic := os.Getenv("MESSAGE_TOPIC")
	kafkaUI := os.Getenv("KAFKA_UI") // Например: "localhost:9093"

	// Превращаем строку брокеров в срез (если хотим поддерживать несколько адресов)
	// Если у вас только один брокер, можно оставить []string{kafkaBrokersStr}
	kafkaBrokers := strings.Split(kafkaBrokersStr, ",")

	return &Config{
		DatabaseURL:  databaseURL,
		KafkaBrokers: kafkaBrokers,
		MessageTopic: messageTopic,
		KafkaUI:      kafkaUI,
	}, nil
}
