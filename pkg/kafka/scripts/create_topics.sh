#!/bin/bash
# Скрипт для создания топика "messages" в Kafka

TOPIC="messages"
BOOTSTRAP_SERVER="localhost:9092"  # Используем внешний слушатель

echo "Проверяем наличие топика '$TOPIC'..."
if docker exec kafka kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER --list | grep -q "^${TOPIC}$"; then
  echo "Топик '$TOPIC' уже существует."
else
  echo "Создаём топик '$TOPIC'..."
  docker exec kafka kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER --create --topic $TOPIC --partitions 1 --replication-factor 1
  if [ $? -eq 0 ]; then
    echo "Топик '$TOPIC' успешно создан."
  else
    echo "Ошибка при создании топика '$TOPIC'."
  fi
fi
