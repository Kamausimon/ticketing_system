package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Consumer wraps a kafka-go Reader.
// Each consumer belongs to a group — Kafka tracks the offset per group,
// so restarting a worker resumes from where it left off.
type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6, // 10 MB
		}),
	}
}

// ReadMessage blocks until the next message is available.
// The caller must call CommitMessage after successfully processing it this advances the offset so the message isn't re-delivered on restart.
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

// CommitMessage marks the message as processed for this consumer group.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
