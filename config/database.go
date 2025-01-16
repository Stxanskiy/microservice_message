package config

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/nevasik7/lg"
	"time"
)

func NewDatabaseConnected(databaseURL string) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DSN ошибка строки плдключения %w", err)
	}
	config.MaxConns = 5
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("Не удалось подключиться к базе данных %w", err)
	}

	//Проверка подклбчения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		lg.Fatalf("База даннвх недоступна %v", err)
		return nil, nil
	}

	lg.Infof("Подключение к базе данных по адресу %s", databaseURL)
	return db, nil
}
