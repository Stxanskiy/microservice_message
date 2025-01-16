package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (kp *KafkaProducer) Publish(ctx context.Context, key string, value []byte) error {
	err := kp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})
	if err != nil {
		log.Printf("Failed to publish message: %v", err)
		return err
	}
	log.Printf("Message published to Kafka: key=%s, value=%s", key, string(value))
	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
