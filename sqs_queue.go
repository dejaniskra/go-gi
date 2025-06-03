package gogi

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSJobQueue struct {
	client   *SQSClient
	queueURL string
}

func NewSQSJobQueue(client *SQSClient, queueURL string) *SQSJobQueue {
	return &SQSJobQueue{client: client, queueURL: queueURL}
}

func (q *SQSJobQueue) SendJob(ctx context.Context, job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = q.client.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.client.QueueURL),
		MessageBody: aws.String(string(data)),
	})
	return err
}

func (q *SQSJobQueue) ReceiveJobs(ctx context.Context, handler func(*Job) error) error {
	resp, err := q.client.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5,
	})
	if err != nil {
		return err
	}

	for _, msg := range resp.Messages {
		var job Job
		if err := json.Unmarshal([]byte(*msg.Body), &job); err != nil {
			return err
		}
		if err := handler(&job); err != nil {
			return err
		}
		_, err = q.client.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(q.queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
