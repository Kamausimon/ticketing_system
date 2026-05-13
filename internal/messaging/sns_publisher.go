package messaging

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// SNSPublisher publishes events to AWS SNS topics.
// It implements the outbox.Sender interface so it is a drop-in replacement
// for the Kafka producer in the outbox publisher and all workers.
type SNSPublisher struct {
	client    *sns.Client
	topicARNs map[string]string // topic name constant → SNS ARN
}

func NewSNSPublisher(client *sns.Client, topicARNs map[string]string) *SNSPublisher {
	return &SNSPublisher{client: client, topicARNs: topicARNs}
}

// Publish sends value to the SNS topic mapped to the given topic name.
// The key is attached as a message attribute so consumers can use it for
// routing or idempotency checks (mirrors the Kafka message key role).
func (p *SNSPublisher) Publish(ctx context.Context, topic, key string, value []byte) error {
	arn, ok := p.topicARNs[topic]
	if !ok {
		return fmt.Errorf("sns_publisher: no ARN configured for topic %q", topic)
	}

	_, err := p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(arn),
		Message:  aws.String(string(value)),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"key": {
				DataType:    aws.String("String"),
				StringValue: aws.String(key),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("sns_publisher: publish to %q: %w", topic, err)
	}
	return nil
}
