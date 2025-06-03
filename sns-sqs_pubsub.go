package gogi

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SNS_SQS_PubSub struct {
	topicARN  string
	queueURL  string
	snsClient *SNSClient
	sqsClient *SQSClient
}

func NewSNS_SQSPubSub(topicARN, queueURL string, snsClient *SNSClient, sqsClient *SQSClient) *SNS_SQS_PubSub {
	return &SNS_SQS_PubSub{
		topicARN:  topicARN,
		queueURL:  queueURL,
		snsClient: snsClient,
		sqsClient: sqsClient,
	}
}

func (p *SNS_SQS_PubSub) Publish(ctx context.Context, data []byte) error {
	_, err := p.snsClient.Client.Publish(ctx, &sns.PublishInput{
		TopicArn: &p.topicARN,
		Message:  aws.String(string(data)),
	})
	return err
}

func (p *SNS_SQS_PubSub) Subscribe(ctx context.Context, handler func([]byte) error) error {
	for {
		output, err := p.sqsClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(p.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     10,
		})
		if err != nil {
			return fmt.Errorf("receive error: %w", err)
		}

		for _, msg := range output.Messages {
			if msg.Body == nil {
				continue
			}

			if err := handler([]byte(*msg.Body)); err != nil {
				log.Printf("handler error: %v", err)
				continue
			}

			_, err := p.sqsClient.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(p.queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("delete message error: %v", err)
			}
		}
	}
}
