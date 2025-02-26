package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"time"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers string) (*Producer, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokers},
		Topic:    "messages",
		Balancer: &kafka.LeastBytes{},
	})
	return &Producer{writer: writer}, nil
}

func (p *Producer) Publish(topic string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msg := kafka.Message{
		Value: message,
	}
	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
