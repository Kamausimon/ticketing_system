package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Message holds the data a worker needs from a received SQS message.
type Message struct {
	// ReceiptHandle is passed to DeleteMessage after successful processing.
	ReceiptHandle string
	// Body is the unwrapped event payload — the inner "Message" field when
	// delivered via SNS fan-out, or the raw body for direct SQS sends.
	Body      []byte
	MessageID string
}

// SQSConsumer receives messages from a single SQS queue using long polling.
// ReceiveMessage maps to Kafka's ReadMessage; DeleteMessage maps to CommitMessage.
type SQSConsumer struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSConsumer(client *sqs.Client, queueURL string) *SQSConsumer {
	return &SQSConsumer{client: client, queueURL: queueURL}
}

// ReceiveMessage blocks for up to 20 seconds (long poll) waiting for a message.
// Returns nil, nil when the poll times out with no messages — callers should loop.
func (c *SQSConsumer) ReceiveMessage(ctx context.Context) (*Message, error) {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, fmt.Errorf("sqs_consumer: receive: %w", err)
	}
	if len(out.Messages) == 0 {
		return nil, nil
	}

	sqsMsg := out.Messages[0]
	body, err := unwrapSNSEnvelope(aws.ToString(sqsMsg.Body))
	if err != nil {
		return nil, fmt.Errorf("sqs_consumer: unwrap envelope: %w", err)
	}

	return &Message{
		ReceiptHandle: aws.ToString(sqsMsg.ReceiptHandle),
		Body:          body,
		MessageID:     aws.ToString(sqsMsg.MessageId),
	}, nil
}

// DeleteMessage removes the message from the queue after successful processing.
func (c *SQSConsumer) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("sqs_consumer: delete: %w", err)
	}
	return nil
}

// snsEnvelope is the JSON wrapper SNS puts around every message it delivers to SQS.
type snsEnvelope struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

// unwrapSNSEnvelope extracts the inner payload from an SNS notification envelope.
// If the body is not an SNS envelope, it is returned as-is.
func unwrapSNSEnvelope(body string) ([]byte, error) {
	var env snsEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		return []byte(body), nil
	}
	if env.Type == "Notification" {
		return []byte(env.Message), nil
	}
	return []byte(body), nil
}
