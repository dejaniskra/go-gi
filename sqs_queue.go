package gogi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSQueue struct {
	client   *sqs.SQS
	queueURL string
}

func NewSQSQueue(sess *session.Session, queueURL string) *SQSQueue {
	return &SQSQueue{
		client:   sqs.New(sess),
		queueURL: queueURL,
	}
}

func (q *SQSQueue) SendJob(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	_, err = q.client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.queueURL),
		MessageBody: aws.String(string(data)),
	})
	return err
}

func (q *SQSQueue) ReceiveJobs(ctx context.Context, handler func(Job) error) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				out, err := q.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(q.queueURL),
					MaxNumberOfMessages: aws.Int64(5),
					WaitTimeSeconds:     aws.Int64(10),
				})
				if err != nil {
					GetLogger().Debug(fmt.Sprintf("SQS receive error: %v", err))
					continue
				}

				for _, msg := range out.Messages {
					var job Job
					if err := json.Unmarshal([]byte(*msg.Body), &job); err != nil {
						GetLogger().Debug(fmt.Sprintf("Failed to unmarshal SQS job: %v", err))
						continue
					}
					if err := handler(job); err != nil {
						GetLogger().Debug(fmt.Sprintf("SQS job handler error: %v", err))
					}
					_, err = q.client.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(q.queueURL),
						ReceiptHandle: msg.ReceiptHandle,
					})
					if err != nil {
						GetLogger().Debug(fmt.Sprintf("Failed to delete SQS message: %v", err))
					}
				}
			}
		}
	}()
	return nil
}
