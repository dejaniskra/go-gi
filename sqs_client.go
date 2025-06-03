package gogi

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	Client   *sqs.Client
	QueueURL string
}

type SQSClientOptions struct {
	AccessKey      string
	SecretKey      string
	CustomEndpoint string
}

type SQLOpt func(*SQSClientOptions)

func WithCredentials(accessKey, secretKey string) SQLOpt {
	return func(opts *SQSClientOptions) {
		opts.AccessKey = accessKey
		opts.SecretKey = secretKey
	}
}

func WithEndpoint(endpoint string) SQLOpt {
	return func(opts *SQSClientOptions) {
		opts.CustomEndpoint = endpoint
	}
}

func GetSQSClient(ctx context.Context, queueName, region string, opts ...SQLOpt) (*SQSClient, error) {
	options := &SQSClientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	configOpts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if options.AccessKey != "" && options.SecretKey != "" {
		configOpts = append(configOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""),
		))
	}

	if options.CustomEndpoint != "" {
		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           options.CustomEndpoint,
				SigningRegion: region,
			}, nil
		})
		configOpts = append(configOpts, config.WithEndpointResolver(customResolver))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	output, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue URL: %w", err)
	}

	return &SQSClient{
		Client:   client,
		QueueURL: *output.QueueUrl,
	}, nil
}
