package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a kafka-go Writer.
// One Producer instance can write to any topic — the topic is set per-message.
type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

// Publish sends a single message to the given topic.
// key is used by Kafka to route the message to a partition — use the entity ID
func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
