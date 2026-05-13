package outbox

import (
	"context"
	"log"
	"time"
)

const (
	maxRetries    = 5
	pollInterval  = time.Second
	batchSize     = 50
)

// Sender is the interface the Publisher uses to send events — keeps the
// publisher decoupled from the concrete Kafka producer.
type Sender interface {
	Publish(ctx context.Context, topic, key string, value []byte) error
}

// Publisher polls the outbox table and forwards pending events to Kafka.
type Publisher struct {
	repo   *Repository
	sender Sender
}

func NewPublisher(repo *Repository, sender Sender) *Publisher {
	return &Publisher{repo: repo, sender: sender}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	log.Println("outbox publisher started")
	for {
		select {
		case <-ctx.Done():
			log.Println("outbox publisher stopped")
			return
		case <-ticker.C:
			p.processBatch(ctx)
		}
	}
}

func (p *Publisher) processBatch(ctx context.Context) {
	events, err := p.repo.FetchPending(batchSize)
	if err != nil {
		log.Printf("outbox: failed to fetch pending events: %v", err)
		return
	}

	for _, evt := range events {
		err := p.sender.Publish(ctx, evt.Topic, evt.Topic, []byte(evt.Payload))
		if err != nil {
			log.Printf("outbox: failed to publish event %d to topic %s: %v", evt.ID, evt.Topic, err)
			p.repo.IncrementRetry(evt.ID, evt.RetryCount+1 >= maxRetries)
			continue
		}

		if err := p.repo.MarkPublished(evt.ID); err != nil {
			// Not fatal — event will be re-published next poll (consumers must be idempotent).
			log.Printf("outbox: published to kafka but failed to mark event %d as done: %v", evt.ID, err)
		}
	}
}
