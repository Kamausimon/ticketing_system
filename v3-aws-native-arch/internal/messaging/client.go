package messaging

import (
	"context"
	"log"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// NewAWSClients loads AWS credentials from the environment and returns
// ready-to-use SNS and SQS clients. Both clients share the same regional config.
func NewAWSClients(ctx context.Context) (*sns.Client, *sqs.Client) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		log.Fatalf("messaging: failed to load AWS config: %v", err)
	}

	// Override endpoint for LocalStack in local development.
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		snsClient := sns.NewFromConfig(cfg, func(o *sns.Options) {
			o.BaseEndpoint = &endpoint
		})
		sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = &endpoint
		})
		return snsClient, sqsClient
	}

	return sns.NewFromConfig(cfg), sqs.NewFromConfig(cfg)
}
