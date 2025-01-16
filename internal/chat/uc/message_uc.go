package uc

import (
	"context"
	"encoding/json"
	"gitlab.com/nevasik7/lg"
	"serice_message/internal/chat/model"
	"serice_message/pkg/kafka"
)

type MessageUC struct {
	kafkaProducer *kafka.KafkaProducer
}

func NewMessageUC(kafkaProducer *kafka.KafkaProducer) *MessageUC {
	return &MessageUC{
		kafkaProducer: kafkaProducer,
	}
}

func (uc *MessageUC) HandleNewMessage(ctx context.Context, msg model.Message) error {
	//Сериализация сообщения в JSON
	data, err := json.Marshal(msg)
	if err != nil {
		lg.Errorf("Failed to serialize %v", err)
		return err
	}

	// Publication message in Kafka
	if err := uc.kafkaProducer.Publish(ctx, string(msg.ChatID), data); err != nil {
		lg.Errorf("fAILED TO PUBLISH MESSAGE %v", msg)
		return err
	}
	//TODO не забыть поменять уровень логирования
	lg.Infof("Message published to Kafka :%v", msg)
	return nil
}
