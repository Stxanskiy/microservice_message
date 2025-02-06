package config

import (
	"github.com/joho/godotenv"
	"gitlab.com/nevasik7/lg"
	"os"
)

type Config struct {
	ServerPort   string
	PostgreURL   string
	CassandraURL string
	KafkaBrokers string
	JwtSecret    string
	JaegerURL    string
}

func LoadCOnfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		lg.Errorf("File .env error to loading: %v", err)
	}

	cfg := &Config{
		ServerPort:   os.Getenv("SERVER_PORT"),
		PostgreURL:   os.Getenv("POSTGRES_URL"),
		CassandraURL: os.Getenv("CASSANDRA_URL"),
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
		JwtSecret:    os.Getenv("JWT_SECRET"),
		JaegerURL:    os.Getenv("JAEGER_URL"),
	}
	return cfg, err
}

/*
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
*/
